package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"interview/internal/httpsvc/domain"
)

func TestRunWithTimeout_ReturnsResult(t *testing.T) {
	got, err := runWithTimeout(context.Background(), time.Second,
		func(context.Context) (domain.Item, error) {
			return domain.Item{ID: "1"}, nil
		})

	require.NoError(t, err)
	assert.Equal(t, "1", got.ID)
}

func TestRunWithTimeout_ExceedsDeadline(t *testing.T) {
	_, err := runWithTimeout(context.Background(), 10*time.Millisecond,
		func(ctx context.Context) (domain.Item, error) {
			<-ctx.Done() // simulate a slow downstream

			return domain.Item{}, ctx.Err()
		})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

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
