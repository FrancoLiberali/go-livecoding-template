// Package controller is the HTTP transport layer. It depends on the
// service.Service interface, so handlers are unit-tested with a mock service.
package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"interview/internal/httpsvc/domain"
	"interview/internal/httpsvc/service"
)

// Controller adapts HTTP requests to the item service.
type Controller struct {
	svc service.Service
}

// New builds a controller over the given service.
func New(svc service.Service) *Controller {
	return &Controller{svc: svc}
}

// RegisterRoutes mounts the controller's routes on the given router.
func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Get("/items/{id}", c.getItem)
	r.Post("/items", c.createItem)
}

type itemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createItemRequest struct {
	Name string `json:"name"`
}

func (c *Controller) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := c.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "item not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, toResponse(item))
}

func (c *Controller) createItem(w http.ResponseWriter, r *http.Request) {
	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")

		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")

		return
	}

	item, err := c.svc.Create(r.Context(), req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusCreated, toResponse(item))
}

func toResponse(item domain.Item) itemResponse {
	return itemResponse{ID: item.ID, Name: item.Name}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
