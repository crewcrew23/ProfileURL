package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"url_profile/internal/app/server/http/constants"
	"url_profile/internal/app/server/http/handlers/viewModel"
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

func (h *ProfileHandler) HandlerMyProfile() http.HandlerFunc {
	type UserView struct {
		Email     string        `json:"email"`
		Username  string        `json:"username"`
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
		uv.Username = u.Username
		uv.AboutText = u.AboutText
		uv.Links = u.Links

		respond(w, http.StatusOK, uv)
	}
}

func (h *ProfileHandler) HandlerGetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]

		if username == "" {
			sendError(w, http.StatusBadRequest, fmt.Errorf("username is required"))
			return
		}

		u, err := h.service.UserByUsername(username)
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

		links := make([]viewModel.LinkView, 0, len(u.Links))
		for _, l := range u.Links {
			links = append(links, viewModel.LinkView{
				LinkName:  l.LinkName,
				LinkColor: l.LinkColor,
				LinkPath:  l.LinkPath,
			})
		}

		uv := &viewModel.UserView{
			Username:  u.Username,
			AboutText: u.AboutText,
			Links:     links,
		}

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
