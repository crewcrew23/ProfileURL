package app

import (
	"log/slog"
	"net/http"
	"url_profile/internal/app/server"
	"url_profile/internal/config"
	authservice "url_profile/internal/services/auth"
	sqlitestore "url_profile/internal/store/sqlite"
)

func Start(cfg config.Config, logger *slog.Logger) error {
	store := sqlitestore.New(cfg.StoragePath, logger)
	authService := authservice.New(logger, store, store)
	srv := server.New(logger, authService)

	return http.ListenAndServe(cfg.Addr, srv.Router) // TODO: configure TLS
}
