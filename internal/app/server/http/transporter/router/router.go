package router

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"url_profile/internal/app/server/http/handlers"
	"url_profile/internal/app/server/http/middleware"
)

func New(
	authHandler *handler.AuthHandlers,
	profileHandler *handler.ProfileHandler,
	linkHandler *handler.LinkHandler,
	log *slog.Logger,
	secret string) *mux.Router {

	r := mux.NewRouter()

	r.Use(middleware.SetRequestID)
	r.Use(middleware.LogRequest(log))
	r.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	//PUBLIC ROUTES
	r.HandleFunc("/api/auth/sign-up", authHandler.HandleSignUp()).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/login", authHandler.HandleLogin()).Methods(http.MethodPost)

	//PUBLIC ROUTES
	public := r.PathPrefix("/api/profile").Subrouter()
	public.HandleFunc("/{username}", profileHandler.HandlerGetProfile()).Methods(http.MethodGet)

	//PRIVATE ROUTES
	private := r.PathPrefix("/api/profile").Subrouter()
	private.Use(middleware.AuthMiddleware(log, secret)) //auth middleware check and verified token
	private.HandleFunc("", profileHandler.HandlerMyProfile()).Methods(http.MethodGet)
	//ABOUT
	private.HandleFunc("/about", profileHandler.HandlerUpdateAboutMe()).Methods(http.MethodPost)
	//lINKS
	private.HandleFunc("/link", linkHandler.HandlerLink()).Methods(http.MethodPost, http.MethodPut, http.MethodDelete)

	return r
}
