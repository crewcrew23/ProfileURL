package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"url_profile/internal/lib/jwt"

	"github.com/google/uuid"
)

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

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header != "" {
			headerAuth := strings.Split(header, " ")
			if len(headerAuth) != 2 || headerAuth[0] != "Bearer" {
				s.error(w, http.StatusUnauthorized, nil)
				return
			}

			claims, err := jwt.ParseAndVerify(headerAuth[1], s.secret)
			if err != nil {
				s.log.Debug("Failed Parse Token: ", slog.String("error", err.Error()))
				s.error(w, http.StatusUnauthorized, nil)
			}

			s.log.Info("token verifIed")
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserIdKey, claims.UID)))
		}
	})
}
