package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Error codes returned in the JSON error envelope.
const (
	codeInvalidRequest  = "invalid_request"
	codeValidationError = "validation_error"
	codeNotFound        = "not_found"
	codeTimeout         = "timeout"
	codeCanceled        = "canceled"
	codeInternal        = "internal"
)

// statusClientClosedRequest is nginx's non-standard 499, used when the client
// disconnects before the server responds (net/http has no constant for it).
const statusClientClosedRequest = 499

// errorResponse is the single JSON envelope used for every error response.
type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []fieldError `json:"details,omitempty"`
}

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// newValidator returns a validator that reports the JSON field name (not the Go
// field name) in errors, so messages line up with the request payload.
func newValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}

		return name
	})

	return validate
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

// writeTransportError renders non-domain (transport/infrastructure) failures.
// Domain errors are mapped per-endpoint by each handler; this covers the
// context-driven cases (client cancellation, deadline) and the catch-all.
// Unexpected (internal) errors are logged in full before the sanitized 500 is
// returned, so the real cause survives while the client sees a generic message.
func writeTransportError(ctx context.Context, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.Canceled):
		// Client went away before we responded; the body is largely moot but we
		// set a status for logs/consistency. 499 = nginx "Client Closed Request".
		writeError(w, statusClientClosedRequest, codeCanceled, "request canceled by client")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, codeTimeout, "request timed out")
	default:
		slog.ErrorContext(ctx, "internal error handling request", "err", err)
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
	}
}

// writeValidationError renders validator failures as a 400 with per-field detail.
func writeValidationError(w http.ResponseWriter, err error) {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		writeError(w, http.StatusBadRequest, codeValidationError, "invalid request")

		return
	}

	details := make([]fieldError, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		details = append(details, fieldError{
			Field:   fieldErr.Field(),
			Message: validationMessage(fieldErr),
		})
	}

	writeJSON(w, http.StatusBadRequest, errorResponse{
		Error: errorBody{
			Code:    codeValidationError,
			Message: "request validation failed",
			Details: details,
		},
	})
}

// validationMessage turns a single validator failure into a human-readable string.
func validationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must be at least " + fieldErr.Param() + " characters"
	case "max":
		return fieldErr.Field() + " must be at most " + fieldErr.Param() + " characters"
	default:
		return fieldErr.Field() + " is invalid"
	}
}
