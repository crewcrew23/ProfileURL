package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"url_profile/internal/app/server/http/handlers/requestModel"
	"url_profile/internal/lib/jwt"
	"url_profile/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandlers struct {
	log      *slog.Logger
	service  UserService
	secret   string
	tokenTTL time.Duration
}

func NewAuthHandlers(log *slog.Logger, service UserService, secret string, tokenTTL time.Duration) *AuthHandlers {
	return &AuthHandlers{
		log:      log,
		service:  service,
		secret:   secret,
		tokenTTL: tokenTTL,
	}
}

func (h *AuthHandlers) HandleSignUp() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		req := &requestModel.SignUpModel{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.log.Debug("Failed to parse", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, err)
			return
		}

		h.log.Debug("ReqUserData:", slog.Any("Data", req))

		if len(req.Password) < 6 {
			sendError(w, http.StatusBadRequest, fmt.Errorf("password cannot be less than 6 characters"))
			return
		}

		if err := req.Validate(); err != nil {
			h.log.Debug("Invalid validate", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, err)
			return
		}

		code, u, err := h.service.CreateUser(req.Email, req.Username, req.Password)
		if err != nil {
			if code == http.StatusConflict {
				sendError(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", req.Email))
				return
			}
			sendError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			return
		}

		h.log.Debug("Created:", slog.Any("data:", u))

		token, err := jwt.NewToken(u, h.tokenTTL, h.secret)
		if err != nil {
			sendError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("token", token)
		respond(w, code, nil)
	}
}

func (h *AuthHandlers) HandleLogin() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		req := &requestModel.LoginModel{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.log.Debug("DECODE ERROR:", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid input data"))
			return
		}

		h.log.Debug("DECODE", slog.Any("data:", req))

		u, err := h.service.User(req.Email)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				sendError(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
				return
			}

			h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			sendError(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		h.log.Debug("User", slog.Any("data", u))

		if err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(req.Password)); err != nil {
			h.log.Debug("Bcryp retrn error from compare password:", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
			return
		}

		token, err := jwt.NewToken(u, h.tokenTTL, h.secret)
		if err != nil {
			h.log.Debug("Error from create jwt:", slog.String("err", err.Error()))
			sendError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			return
		}

		w.Header().Set("token", token)
		respond(w, http.StatusOK, nil)
	}
}
