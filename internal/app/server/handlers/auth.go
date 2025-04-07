package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"
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
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			h.log.Debug("Failed to parse", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, err)
			return
		}

		h.log.Debug("ReqUserData:", slog.Any("Data", req))

		if req.Email == "" {
			sendError(w, http.StatusBadRequest, fmt.Errorf("email is required"))
			return
		}

		if len(req.Password) < 6 {
			sendError(w, http.StatusBadRequest, fmt.Errorf("password cannot be less than 6 characters"))
			return
		}

		if !isValidEmail(req.Email) {
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid email"))
			return
		}

		code, u, err := h.service.CreateUser(req.Email, req.Password)
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
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

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

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}
