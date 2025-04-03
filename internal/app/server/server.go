package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"url_profile/internal/domain/models"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type ctxKey int8

const (
	ctxRequestKey ctxKey = iota
)

type UserService interface {
	CreateUser(email string, password string) (code int, err error)
	User(email string) (*models.User, error)
}

type server struct {
	log         *slog.Logger
	Router      *mux.Router
	userService UserService
	// TODO: storage
}

func New(log *slog.Logger, userService UserService) *server {
	s := &server{
		log:         log,
		Router:      mux.NewRouter(),
		userService: userService,
	}

	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.Router.Use(s.setRequetID)
	s.Router.Use(s.logRequest)
	s.Router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.Router.HandleFunc("/api/auth/sign-up", s.handleSignUp()).Methods(http.MethodPost)
	s.Router.HandleFunc("/api/profile", s.handleProfile()).Methods(http.MethodGet)
	s.Router.HandleFunc("/api/", s.checkHandler()).Methods(http.MethodGet)
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
		}

		s.log.Debug("ReqUserData:", slog.Any("Data", req))

		code, err := s.userService.CreateUser(req.Email, req.Password)
		if err != nil {
			s.error(w, code, err)
		}

		s.respond(w, code, nil)
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
		}

		u, err := s.userService.User(req.Email)
		if err != nil {
			s.error(w, 404, err)
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
	s.respond(w, code, err)
}

func (s *server) respond(w http.ResponseWriter, code int, data interface{}) {
	if data != nil {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
	w.WriteHeader(code)
}
