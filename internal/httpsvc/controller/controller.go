// Package controller is the HTTP transport layer. It depends on the
// service.Service interface, so handlers are unit-tested with a mock service.
// Every response — success or error — is JSON. Request bodies are validated
// with go-playground/validator, and each endpoint runs under its own timeout
// that yields a 504 when exceeded.
package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"interview/internal/httpsvc/domain"
	"interview/internal/httpsvc/service"
)

const (
	defaultGetItemTimeout    = 2 * time.Second
	defaultCreateItemTimeout = 3 * time.Second
)

// Timeouts holds the per-endpoint request timeouts.
type Timeouts struct {
	GetItem    time.Duration
	CreateItem time.Duration
}

// DefaultTimeouts returns sensible per-endpoint defaults.
func DefaultTimeouts() Timeouts {
	return Timeouts{
		GetItem:    defaultGetItemTimeout,
		CreateItem: defaultCreateItemTimeout,
	}
}

// Controller adapts HTTP requests to the item service.
type Controller struct {
	svc      service.Service
	validate *validator.Validate
	timeouts Timeouts
}

// New builds a controller with the default per-endpoint timeouts.
func New(svc service.Service) *Controller {
	return NewWithTimeouts(svc, DefaultTimeouts())
}

// NewWithTimeouts builds a controller with explicit timeouts (used in tests).
func NewWithTimeouts(svc service.Service, timeouts Timeouts) *Controller {
	return &Controller{svc: svc, validate: newValidator(), timeouts: timeouts}
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
	Name string `json:"name" validate:"required,max=100"`
}

func (c *Controller) getItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	item, err := runWithTimeout(r.Context(), c.timeouts.GetItem,
		func(ctx context.Context) (domain.Item, error) {
			return c.svc.Get(ctx, id)
		})
	if err != nil {
		c.writeServiceError(w, err)

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

	item, err := runWithTimeout(r.Context(), c.timeouts.CreateItem,
		func(ctx context.Context) (domain.Item, error) {
			return c.svc.Create(ctx, req.Name)
		})
	if err != nil {
		c.writeServiceError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, toResponse(item))
}

// writeServiceError maps a service-layer error to the right status + JSON envelope.
func (c *Controller) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, codeTimeout, "request timed out")
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, codeNotFound, "item not found")
	default:
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
	}
}

func toResponse(item domain.Item) itemResponse {
	return itemResponse{ID: item.ID, Name: item.Name}
}

// runWithTimeout runs op under a deadline. If the deadline elapses before op
// returns, it returns context.DeadlineExceeded regardless of whether op honors
// the context — guaranteeing the caller can surface a 504. op still receives
// the deadline-bound context so well-behaved downstreams cancel promptly.
func runWithTimeout[T any](
	parent context.Context,
	timeout time.Duration,
	op func(context.Context) (T, error),
) (T, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	type result struct {
		val T
		err error
	}

	// Buffered so the goroutine never blocks if we've already timed out.
	done := make(chan result, 1)

	go func() {
		val, err := op(ctx)
		done <- result{val: val, err: err}
	}()

	select {
	case <-ctx.Done():
		var zero T

		return zero, ctx.Err()
	case res := <-done:
		return res.val, res.err
	}
}
