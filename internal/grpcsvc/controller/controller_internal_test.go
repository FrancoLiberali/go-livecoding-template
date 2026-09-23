package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/domain"
)

func TestRunWithTimeout_ExceedsDeadline(t *testing.T) {
	_, err := runWithTimeout(context.Background(), 10*time.Millisecond,
		func(ctx context.Context) (domain.Item, error) {
			<-ctx.Done() // simulate a slow downstream

			return domain.Item{}, ctx.Err()
		})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestTransportError_Timeout(t *testing.T) {
	err := transportError(context.DeadlineExceeded)

	require.Error(t, err)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

func TestTransportError_Internal(t *testing.T) {
	err := transportError(errors.New("boom"))

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
