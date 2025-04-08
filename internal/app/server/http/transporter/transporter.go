package transporter

import (
	"github.com/gorilla/mux"
	"log/slog"
	"time"
	"url_profile/internal/app/server/http/handlers"
	serviceinterface "url_profile/internal/app/server/http/transporter/interfaces/service"
	"url_profile/internal/app/server/http/transporter/router"
)

func NewRouter(log *slog.Logger, userService serviceinterface.UserService, secret string, tokenTTL time.Duration) *mux.Router {
	authHandler := handler.NewAuthHandlers(log, userService, secret, tokenTTL)
	profileHandler := handler.NewProfileHandlers(log, userService)
	linkHandler := handler.NewLinkHandlers(log, userService)

	return router.New(authHandler, profileHandler, linkHandler, log, secret)
}
