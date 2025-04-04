package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"url_profile/internal/domain/models"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type ctxKey int8

const (
	ctxRequestKey ctxKey = iota
	ctxUserIdKey  ctxKey = iota
)

type UserService interface {
	CreateUser(email string, password string) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
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
		tokenTTL:    tokenTTL,
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
	public := s.Router.PathPrefix("/api/profile").Subrouter()
	private := s.Router.PathPrefix("/api/profile").Subrouter()

	public.HandleFunc("/{email}", s.handlerGetProfile()).Methods(http.MethodGet)
	private.Use(s.authMiddleware)
	private.HandleFunc("", s.handlerMyProfle()).Methods(http.MethodGet)
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
