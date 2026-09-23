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

func TestUnaryLogger_LogsMethodAndCode(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	interceptor := grpcmw.UnaryLogger(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/grpcsvc.v1.ItemService/GetItem"}
	handler := func(context.Context, any) (any, error) {
		return nil, status.Error(codes.NotFound, "nope")
	}

	_, err := interceptor(context.Background(), nil, info, handler)

	require.Error(t, err)

	line := buf.String()
	assert.Contains(t, line, `"msg":"grpc_request"`)
	assert.Contains(t, line, `"method":"/grpcsvc.v1.ItemService/GetItem"`)
	assert.Contains(t, line, `"code":"NotFound"`)
}
