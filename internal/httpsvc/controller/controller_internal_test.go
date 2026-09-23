package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTransportError_Timeout(t *testing.T) {
	rec := httptest.NewRecorder()

	writeTransportError(rec, context.DeadlineExceeded)

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"timeout","message":"request timed out"}}`, rec.Body.String())
}

func TestWriteTransportError_Internal(t *testing.T) {
	rec := httptest.NewRecorder()

	writeTransportError(rec, errors.New("boom"))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"internal","message":"internal error"}}`, rec.Body.String())
}
