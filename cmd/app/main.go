package main

import (
	"log/slog"
	"url_profile/internal/app"
	"url_profile/internal/app/logger"
	"url_profile/internal/config"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetUpLogger(cfg.Env)

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Any("cfg", cfg),
		slog.String("port", cfg.Addr),
	)

	if err := app.Start(*cfg, log); err != nil {
		panic(err)
	}

}
