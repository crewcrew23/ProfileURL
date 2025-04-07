package middleware

import (
	"context"
	"encoding/json"
	"errors"
	jwt_go "github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"strings"
	"time"
	consts "url_profile/internal/app/server/constants"
	"url_profile/internal/lib/jwt"

	"github.com/google/uuid"
)

func SetRequetID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-REQUEST-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), consts.CtxRequestKey, id)))
	})
}

func LogRequest(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.With(
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.String("user_agent", r.UserAgent()),
			)

			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}

			next.ServeHTTP(rw, r)

			logger.Info(
				"request completed",
				slog.Int("status", rw.code),
				slog.String("duration", time.Since(start).String()),
			)
		})
	}
}

func AuthMiddleware(log *slog.Logger, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			log.Info("header", slog.Any("AUTH", header))

			if header == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			headerAuth := strings.Split(header, " ")
			if len(headerAuth) != 2 || headerAuth[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ParseAndVerify(headerAuth[1], secret)
			if err != nil {

				if errors.Is(err, jwt_go.ErrTokenExpired) {
					log.Debug("Token expired")
					json.NewEncoder(w).Encode(map[string]string{
						"error": "token expired",
					})
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				log.Debug("Failed Parse Token: ", slog.String("error", err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			log.Info("token verified")
			ctx := context.WithValue(r.Context(), consts.CtxUserIdKey, claims.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
