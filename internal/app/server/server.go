package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"url_profile/internal/domain/models"
	"url_profile/internal/lib/jwt"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type ctxKey int8

const (
	ctxRequestKey ctxKey = iota
)

type UserService interface {
	CreateUser(email string, password string) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
}

type server struct {
	log         *slog.Logger
	Router      *mux.Router
	userService UserService
	secret      string
	tokenTTL    time.Duration
}

func New(log *slog.Logger, userService UserService, secret string, tokenTTL time.Duration) *server {
	s := &server{
		log:         log,
		Router:      mux.NewRouter(),
		userService: userService,
		secret:      secret,
	}

	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.Router.Use(s.setRequetID)
	s.Router.Use(s.logRequest)
	s.Router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.Router.HandleFunc("/api/auth/sign-up", s.handleSignUp()).Methods(http.MethodPost)
	s.Router.HandleFunc("/api/auth/login", s.handleLogin()).Methods(http.MethodPost)
	s.Router.HandleFunc("/health", s.checkHandler()).Methods(http.MethodGet)

	// TODO:
	// public := s.Router.PathPrefix("/api/profile").Subrouter()
	// private := s.Router.PathPrefix("/api/profile").Subrouter()

	// public.HandleFunc("/{id}")
	// private.Use(auth_middleware)
	// private.HandleFunc("", getMyProfleHandler).Methods(http.MethodGet)
}

func (s *server) setRequetID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-REQUEST-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxRequestKey, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.log.With(
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("user_agent", r.UserAgent()),
		)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		logger.Info(
			"completed with",
			slog.String("Status code", fmt.Sprintf("%d %s", rw.code, http.StatusText(rw.code))),
			slog.Any("Time", time.Since(start)),
		)

	})
}

func (s *server) handleSignUp() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, http.StatusBadRequest, err)
			return
		}

		s.log.Debug("ReqUserData:", slog.Any("Data", req))

		code, u, err := s.userService.CreateUser(req.Email, req.Password)
		if err != nil {
			if code == http.StatusConflict {
				s.error(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", req.Email))
				return
			}
			s.error(w, code, err)
			return
		}

		s.log.Debug("Created:", slog.Any("data:", u))

		token, err := jwt.NewToken(u, s.tokenTTL, s.secret)
		if err != nil {
			s.error(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("token", token)
		s.respond(w, code, nil)
	}
}

func (s *server) handleLogin() http.HandlerFunc {

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.log.Debug("DECODE ERROR:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid input data"))
			return
		}

		u, err := s.userService.User(req.Email)
		if err != nil {
			s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
			return
		}

		if err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(req.Password)); err != nil {
			s.log.Debug("Bcryp retrn error from compare password:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
			return
		}

		token, err := jwt.NewToken(u, s.tokenTTL, s.secret)
		if err != nil {
			s.log.Debug("Error from create jwt:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			return
		}

		w.Header().Set("token", token)
		s.respond(w, http.StatusOK, nil)
	}
}

func (s *server) handleProfile() http.HandlerFunc {

	type request struct {
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, http.StatusBadRequest, err)
			return
		}

		u, err := s.userService.User(req.Email)
		if err != nil {
			s.error(w, 404, err)
			return
		}

		s.respond(w, 200, u)
	}
}

func (s *server) checkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, 200, nil)
	}
}

func (s *server) error(w http.ResponseWriter, code int, err error) {
	s.respond(w, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, code int, data interface{}) {
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
	}

	w.WriteHeader(code)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
