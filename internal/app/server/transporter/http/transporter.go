package http

import (
	"github.com/gorilla/mux"
	"log/slog"
	"time"
	handler "url_profile/internal/app/server/handlers"
	"url_profile/internal/app/server/router"
	serviceinterface "url_profile/internal/app/server/transporter/interfaces/service"
)

func NewRouter(log *slog.Logger, userService serviceinterface.UserService, secret string, tokenTTL time.Duration) *mux.Router {
	authHandler := handler.NewAuthHandlers(log, userService, secret, tokenTTL)
	profileHandler := handler.NewProfileHandlers(log, userService)
	linkHandler := handler.NewLinkHandlers(log, userService)

	return router.New(authHandler, profileHandler, linkHandler, log, secret)
}
