package grpcmw_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/grpcmw"
)

func TestUnaryRecovery_RecoversPanic(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	interceptor := grpcmw.UnaryRecovery(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/grpcsvc.v1.ItemService/GetItem"}
	handler := func(context.Context, any) (any, error) {
		panic("boom")
	}

	resp, err := interceptor(context.Background(), nil, info, handler)

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, buf.String(), "boom")
}
