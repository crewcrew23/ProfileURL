package server

import (
	"log/slog"
	"net/http"
	"time"
	handler "url_profile/internal/app/server/handlers"
	"url_profile/internal/app/server/middleware"
	"url_profile/internal/domain/models"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type UserService interface {
	CreateUser(email string, password string) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
	UpdateAboutMe(id int, text string) error
	AddLink(userID int, link models.ReqLink) error
	UpdateLink(userID int, link *models.ReqUpdateLink) error
	DeleteLink(userID int, linkID int) error
}

type server struct {
	log            *slog.Logger
	Router         *mux.Router
	userService    UserService
	secret         string
	authHandler    *handler.AuthHandlers
	profileHandler *handler.ProfileHandler
	linkHandler    *handler.LinkHandler
}

func New(log *slog.Logger, userService UserService, secret string, tokenTTL time.Duration) *server {
	authHandler := handler.NewAuthHandlers(log, userService, secret, tokenTTL)
	profileHandler := handler.NewProfileHandlers(log, userService)
	linkHandler := handler.NewLinkHandlers(log, userService)

	s := &server{
		log:            log,
		Router:         mux.NewRouter(),
		secret:         secret,
		userService:    userService,
		authHandler:    authHandler,
		profileHandler: profileHandler,
		linkHandler:    linkHandler,
	}

	s.configureRouter()
	return s
}

func (s *server) configureRouter() {
	s.Router.Use(middleware.SetRequetID)
	s.Router.Use(middleware.LogRequest(s.log))
	s.Router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	//PUBLIC ROUTES
	s.Router.HandleFunc("/api/auth/sign-up", s.authHandler.HandleSignUp()).Methods(http.MethodPost)
	s.Router.HandleFunc("/api/auth/login", s.authHandler.HandleLogin()).Methods(http.MethodPost)

	//PUBLIC ROUTES
	public := s.Router.PathPrefix("/api/profile").Subrouter()
	public.HandleFunc("/{email}", s.profileHandler.HandlerGetProfile()).Methods(http.MethodGet)

	//PRIVATE ROUTES
	private := s.Router.PathPrefix("/api/profile").Subrouter()
	private.Use(middleware.AuthMiddleware(s.log, s.secret)) //auth middleware check and verified token
	private.HandleFunc("", s.profileHandler.HandlerMyProfle()).Methods(http.MethodGet)
	//ABOUT
	private.HandleFunc("/about", s.profileHandler.HandlerUpdateAboutMe()).Methods(http.MethodPost)
	//lINKS
	private.HandleFunc("/link", s.linkHandler.HandlerLink()).Methods(http.MethodPost, http.MethodPut, http.MethodDelete)
}
