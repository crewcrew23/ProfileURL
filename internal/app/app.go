package app

import (
	"log/slog"
	"net/http"
	"time"
	"url_profile/internal/app/server"
	"url_profile/internal/config"
	authservice "url_profile/internal/services/auth"
	sqlitestore "url_profile/internal/store/sqlite"
)

func Start(cfg config.Config, logger *slog.Logger) error {
	store := sqlitestore.New(cfg.StoragePath, logger)
	authService := authservice.New(logger, store, store)
	duration, err := time.ParseDuration(cfg.TokenTLL)
	if err != nil {
		panic(err)
	}
	srv := server.New(logger, authService, cfg.Secret, duration)

	return http.ListenAndServe(cfg.Addr, srv.Router) // TODO: configure TLS
}
