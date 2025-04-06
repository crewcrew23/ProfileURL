package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	consts "url_profile/internal/app/server/constants"
	"url_profile/internal/domain/models"
	"url_profile/internal/store"

	"github.com/gorilla/mux"
)

type ProfileHandler struct {
	log     *slog.Logger
	service UserService
}

func NewProfileHandlers(log *slog.Logger, service UserService) *ProfileHandler {
	return &ProfileHandler{
		log:     log,
		service: service,
	}
}

func (h *ProfileHandler) HandlerMyProfle() http.HandlerFunc {
	type UserView struct {
		Email     string        `json:"email"`
		AboutText string        `json:"about"`
		Links     []models.Link `json:"links"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		uv := &UserView{}
		h.log.Debug("UserID in Context: ", slog.Int("ctxID", r.Context().Value(consts.CtxUserIdKey).(int)))
		u, err := h.service.UserById(r.Context().Value(consts.CtxUserIdKey).(int))
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				sendError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
				return
			}

			h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			sendError(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		h.log.Debug("User", slog.Any("data", u))

		uv.Email = u.Email
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		respond(w, http.StatusOK, uv)
	}
}

func (h *ProfileHandler) HandlerGetProfile() http.HandlerFunc {

	type UserView struct {
		Email     string        `json:"email"`
		AboutText string        `json:"about"`
		Links     []models.Link `json:"links"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		email := vars["email"]

		if email == "" {
			sendError(w, http.StatusBadRequest, fmt.Errorf("email is required"))
			return
		}

		u, err := h.service.User(email)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
				sendError(w, http.StatusBadRequest, fmt.Errorf("user not found"))
				return
			}

			h.log.Debug("Find User Return Error:", slog.String("err", err.Error()))
			sendError(w, http.StatusInternalServerError, fmt.Errorf("server internal error"))
			return
		}

		uv := &UserView{}

		uv.Email = u.Email
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		respond(w, http.StatusOK, uv)
	}
}

func (h *ProfileHandler) HandlerUpdateAboutMe() http.HandlerFunc {
	type ReqText struct {
		Text string `json:"text"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		text := &ReqText{}
		if err := json.NewDecoder(r.Body).Decode(text); err != nil {
			h.log.Debug("DECODE ERROR:", slog.String("err", err.Error()))
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid input data"))
			return
		}

		if err := h.service.UpdateAboutMe(r.Context().Value(consts.CtxUserIdKey).(int), text.Text); err != nil {
			if errors.Is(err, store.ErrNoRowsAffected) {
				h.log.Debug("DataBase Error:", slog.String("err", err.Error()))
				sendError(w, http.StatusConflict, err)
			}
			h.log.Debug("DataBase Error:", slog.String("err", err.Error()))
			sendError(w, http.StatusInternalServerError, err)
			return
		}

		respond(w, http.StatusOK, nil)
	}
}
