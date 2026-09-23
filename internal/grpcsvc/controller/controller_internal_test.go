package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTransportError_Timeout(t *testing.T) {
	err := transportError(context.DeadlineExceeded)

	require.Error(t, err)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

func TestTransportError_Canceled(t *testing.T) {
	err := transportError(context.Canceled)

	require.Error(t, err)
	assert.Equal(t, codes.Canceled, status.Code(err))
}

func TestTransportError_Internal(t *testing.T) {
	err := transportError(errors.New("boom"))

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
