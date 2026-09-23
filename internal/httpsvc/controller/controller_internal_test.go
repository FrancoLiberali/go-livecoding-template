package controller

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTransportError_Timeout(t *testing.T) {
	rec := httptest.NewRecorder()

	writeTransportError(context.Background(), rec, context.DeadlineExceeded)

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"timeout","message":"request timed out"}}`, rec.Body.String())
}

func TestWriteTransportError_Canceled(t *testing.T) {
	rec := httptest.NewRecorder()

	writeTransportError(context.Background(), rec, context.Canceled)

	require.Equal(t, 499, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"canceled","message":"request canceled by client"}}`, rec.Body.String())
}

func TestWriteTransportError_Internal(t *testing.T) {
	var logs bytes.Buffer

	prev := slog.Default()
	defer slog.SetDefault(prev)

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))

	rec := httptest.NewRecorder()

	writeTransportError(context.Background(), rec, errors.New("db exploded"))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	// Sanitized response to the client...
	assert.JSONEq(t, `{"error":{"code":"internal","message":"internal error"}}`, rec.Body.String())
	// ...but the real cause is logged.
	assert.Contains(t, logs.String(), "db exploded")
	assert.NotContains(t, rec.Body.String(), "db exploded")
}
