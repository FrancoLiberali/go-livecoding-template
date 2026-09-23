package controller

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTransportError_Timeout(t *testing.T) {
	err := transportError(context.Background(), context.DeadlineExceeded)

	require.Error(t, err)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

func TestTransportError_Canceled(t *testing.T) {
	err := transportError(context.Background(), context.Canceled)

	require.Error(t, err)
	assert.Equal(t, codes.Canceled, status.Code(err))
}

func TestTransportError_Internal(t *testing.T) {
	var logs bytes.Buffer

	prev := slog.Default()
	defer slog.SetDefault(prev)

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))

	err := transportError(context.Background(), errors.New("db exploded"))

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	// Sanitized status message to the client...
	assert.Equal(t, "internal error", status.Convert(err).Message())
	// ...but the real cause is logged.
	assert.Contains(t, logs.String(), "db exploded")
}
