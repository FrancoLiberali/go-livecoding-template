// Package controller is the HTTP transport layer. It depends on the
// service.Service interface, so handlers are unit-tested with a mock service.
// Every response — success or error — is JSON. Request bodies are validated
// with go-playground/validator. Each endpoint gets a deadline from a per-route
// chi timeout middleware; when the service honors that deadline and returns
// context.DeadlineExceeded, the handler surfaces a 504.
package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"interview/internal/httpsvc/domain"
	"interview/internal/httpsvc/service"
)

// Per-endpoint request timeouts.
const (
	getItemTimeout    = 2 * time.Second
	createItemTimeout = 3 * time.Second
)

// Controller adapts HTTP requests to the item service.
type Controller struct {
	svc      service.Service
	validate *validator.Validate
}

// New builds a controller over the given service.
func New(svc service.Service) *Controller {
	return &Controller{svc: svc, validate: newValidator()}
}

// RegisterRoutes mounts the routes, each with its own deadline middleware.
func (c *Controller) RegisterRoutes(r chi.Router) {
	r.With(middleware.Timeout(getItemTimeout)).Get("/items/{id}", c.getItem)
	r.With(middleware.Timeout(createItemTimeout)).Post("/items", c.createItem)
}

type itemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createItemRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (c *Controller) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := c.svc.Get(r.Context(), id)
	if err != nil {
		// This endpoint owns how its domain errors map to responses.
		switch {
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, http.StatusNotFound, codeNotFound, "item not found")
		default:
			writeTransportError(r.Context(), w, err)
		}

		return
	}

	writeJSON(w, http.StatusOK, toResponse(item))
}

func (c *Controller) createItem(w http.ResponseWriter, r *http.Request) {
	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, codeInvalidRequest, "malformed JSON body")

		return
	}

	if err := c.validate.Struct(req); err != nil {
		writeValidationError(w, err)

		return
	}

	item, err := c.svc.Create(r.Context(), req.Name)
	if err != nil {
		// No domain errors are expected from Create; only transport-level ones.
		writeTransportError(r.Context(), w, err)

		return
	}

	writeJSON(w, http.StatusCreated, toResponse(item))
}

func toResponse(item domain.Item) itemResponse {
	return itemResponse{ID: item.ID, Name: item.Name}
}
