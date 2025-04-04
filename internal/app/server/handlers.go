package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"url_profile/internal/domain/models"
	"url_profile/internal/lib/jwt"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func (s *server) handleSignUp() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, http.StatusBadRequest, err)
			return
		}

		s.log.Debug("ReqUserData:", slog.Any("Data", req))

		code, u, err := s.userService.CreateUser(req.Email, req.Password)
		if err != nil {
			if code == http.StatusConflict {
				s.error(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", req.Email))
				return
			}
			s.error(w, code, err)
			return
		}

		s.log.Debug("Created:", slog.Any("data:", u))

		token, err := jwt.NewToken(u, s.tokenTTL, s.secret)
		if err != nil {
			s.error(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("token", token)
		s.respond(w, code, nil)
	}
}

func (s *server) handleLogin() http.HandlerFunc {

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.log.Debug("DECODE ERROR:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid input data"))
			return
		}

		u, err := s.userService.User(req.Email)
		if err != nil {
			s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
			return
		}

		if err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(req.Password)); err != nil {
			s.log.Debug("Bcryp retrn error from compare password:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
			return
		}

		token, err := jwt.NewToken(u, s.tokenTTL, s.secret)
		if err != nil {
			s.log.Debug("Error from create jwt:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			return
		}

		w.Header().Set("token", token)
		s.respond(w, http.StatusOK, nil)
	}
}

func (s *server) handlerMyProfle() http.HandlerFunc {
	type UserView struct {
		Email     string        `json:"email"`
		AboutText string        `json:"about"`
		Links     []models.Link `json:"links"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uv := &UserView{}
		s.log.Debug("UserID in Context: ", slog.Int("ctxID", r.Context().Value(ctxUserIdKey).(int)))
		u, err := s.userService.UserById(r.Context().Value(ctxUserIdKey).(int))
		if err != nil {
			s.error(w, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		uv.Email = u.Email
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		s.respond(w, http.StatusOK, uv)
	}
}

func (s *server) handlerGetProfile() http.HandlerFunc {

	type UserView struct {
		Email     string        `json:"email"`
		AboutText string        `json:"about"`
		Links     []models.Link `json:"links"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		email := vars["email"]

		if email == "" {
			s.error(w, http.StatusBadRequest, fmt.Errorf("email is required"))
			return
		}

		u, err := s.userService.User(email)
		if err != nil {
			s.error(w, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		uv := &UserView{}

		uv.Email = u.Email
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		s.respond(w, http.StatusOK, uv)
	}
}

func (s *server) checkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, 200, nil)
	}
}
