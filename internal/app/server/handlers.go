package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"url_profile/internal/domain/models"
	"url_profile/internal/lib/jwt"
	"url_profile/internal/store"

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
			s.error(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
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
			if errors.Is(err, store.ErrUserNotFound) {
				s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				s.error(w, http.StatusBadRequest, fmt.Errorf("incorrect login or password"))
				return
			}

			s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		s.log.Debug("User", slog.Any("data", u))

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
			if errors.Is(err, store.ErrUserNotFound) {
				s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				s.error(w, http.StatusBadRequest, fmt.Errorf("user not found"))
				return
			}

			s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		s.log.Debug("User", slog.Any("data", u))

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
			if errors.Is(err, store.ErrUserNotFound) {
				s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				s.error(w, http.StatusBadRequest, fmt.Errorf("user not found"))
				return
			}

			s.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		uv := &UserView{}

		uv.Email = u.Email
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		s.respond(w, http.StatusOK, uv)
	}
}

func (s *server) handlerUpdateAboutMe() http.HandlerFunc {
	type ReqText struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		text := &ReqText{}
		if err := json.NewDecoder(r.Body).Decode(text); err != nil {
			s.log.Debug("DECODE ERROR:", slog.String("err", err.Error()))
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid input data"))
			return
		}

		if err := s.userService.UpdateAboutMe(r.Context().Value(ctxUserIdKey).(int), text.Text); err != nil {
			if errors.Is(err, store.ErrNoRowsAffected) {
				s.log.Debug("DataBase Error:", slog.String("err", err.Error()))
				s.error(w, http.StatusConflict, err)
			}
			s.log.Debug("DataBase Error:", slog.String("err", err.Error()))
			s.error(w, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, http.StatusOK, nil)
	}
}

func (s *server) checkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, 200, nil)
	}
}

func (s *server) handlerAddLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxUserIdKey).(int)
		var links []models.ReqLink
		if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		for _, link := range links {
			if link.LinkName == "" || link.LinkPath == "" {
				s.error(w, http.StatusBadRequest, fmt.Errorf("link_name and link_path are required"))
				return
			}

			if err := s.userService.AddLink(userID, link); err != nil {
				if errors.Is(err, store.ErrLinkAlreadyExists) {
					s.error(w, http.StatusConflict, err)
					return
				}

				if errors.Is(err, store.ErrUserNotFound) {
					s.error(w, http.StatusNotFound, err)
					return
				}

				s.log.Debug("Error Create Link", slog.String("error", err.Error()))
				s.error(w, http.StatusInternalServerError, err)
				return
			}
		}

		s.respond(w, http.StatusOK, nil)
	}
}

func (s *server) handlerUpdateLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxUserIdKey).(int)
		link := &models.ReqUpdateLink{}
		if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		if err := s.userService.UpdateLink(userID, link); err != nil {
			if errors.Is(err, store.ErrLinkNotFound) {
				s.error(w, http.StatusNotFound, nil)
				return
			}
			s.error(w, http.StatusInternalServerError, nil)
		}

		s.respond(w, http.StatusOK, nil)
	}

}

func (s *server) handlerDeleteLink() http.HandlerFunc {
	type LinkID struct {
		LinkID int `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxUserIdKey).(int)
		link := &LinkID{}
		if err := json.NewDecoder(r.Body).Decode(link); err != nil {
			s.error(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		if err := s.userService.DeleteLink(userID, link.LinkID); err != nil {
			if errors.Is(err, store.ErrLinkNotFound) {
				s.error(w, http.StatusNotFound, err)
				return
			}
			s.error(w, http.StatusInternalServerError, nil)
		}

		s.respond(w, http.StatusOK, nil)
	}
}
