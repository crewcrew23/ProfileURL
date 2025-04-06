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
)

type LinkHandler struct {
	log     *slog.Logger
	service UserService
}

func NewLinkHandlers(log *slog.Logger, service UserService) *LinkHandler {
	return &LinkHandler{
		log:     log,
		service: service,
	}
}

func (h *LinkHandler) HandlerLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.handlerAddLink()(w, r)
		case http.MethodPut:
			h.handlerUpdateLink()(w, r)
		case http.MethodDelete:
			h.handlerDeleteLink()(w, r)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}

func (s *LinkHandler) handlerAddLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(consts.CtxUserIdKey).(int)
		var links []models.ReqLink
		if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		for _, link := range links {
			if link.LinkName == "" || link.LinkPath == "" {
				sendError(w, http.StatusBadRequest, fmt.Errorf("link_name and link_path are required"))
				return
			}

			if err := s.service.AddLink(userID, link); err != nil {
				if errors.Is(err, store.ErrLinkAlreadyExists) {
					sendError(w, http.StatusConflict, err)
					return
				}

				if errors.Is(err, store.ErrUserNotFound) {
					sendError(w, http.StatusNotFound, err)
					return
				}

				s.log.Debug("Error Create Link", slog.String("error", err.Error()))
				sendError(w, http.StatusInternalServerError, err)
				return
			}
		}

		respond(w, http.StatusOK, nil)
	}
}

func (h *LinkHandler) handlerUpdateLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(consts.CtxUserIdKey).(int)
		link := &models.ReqUpdateLink{}
		if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		if err := h.service.UpdateLink(userID, link); err != nil {
			if errors.Is(err, store.ErrLinkNotFound) {
				sendError(w, http.StatusNotFound, fmt.Errorf(""))
				return
			}
			sendError(w, http.StatusInternalServerError, fmt.Errorf(""))
		}

		respond(w, http.StatusOK, nil)
	}

}

func (h *LinkHandler) handlerDeleteLink() http.HandlerFunc {
	type LinkID struct {
		LinkID int `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(consts.CtxUserIdKey).(int)
		link := &LinkID{}
		if err := json.NewDecoder(r.Body).Decode(link); err != nil {
			sendError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
			return
		}

		if err := h.service.DeleteLink(userID, link.LinkID); err != nil {
			if errors.Is(err, store.ErrLinkNotFound) {
				sendError(w, http.StatusNotFound, err)
				return
			}
			sendError(w, http.StatusInternalServerError, fmt.Errorf(""))
		}

		respond(w, http.StatusOK, nil)
	}
}
