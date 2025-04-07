package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	transport "url_profile/internal/app/server/transporter/http"
	"url_profile/internal/config"
	authservice "url_profile/internal/services/auth"
	sqlitestore "url_profile/internal/store/sqlite"
)

func Start(cfg config.Config, logger *slog.Logger) error {
	store := sqlitestore.New(cfg.StoragePath, logger)
	authService := authservice.New(logger, store, store)
	duration, err := time.ParseDuration(cfg.TokenTTL)
	if err != nil {
		panic(fmt.Errorf("failed to parse TokenTTL: %w", err))
	}
	fmt.Printf("Parsed duration: %v\n", duration)
	router := transport.NewRouter(logger, authService, cfg.Secret, duration)

	return http.ListenAndServe(cfg.Addr, router) // TODO: configure TLS
}
