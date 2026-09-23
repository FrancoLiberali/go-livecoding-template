// Package grpcmw holds gRPC server interceptors for the gRPC service.
package grpcmw

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryLogger logs one structured line per unary RPC, capturing the method, the
// resulting status code, and the latency.
func UnaryLogger(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logger.LogAttrs(ctx, slog.LevelInfo, "grpc_request",
			slog.String("method", info.FullMethod),
			slog.String("code", status.Code(err).String()),
			slog.Duration("duration", time.Since(start)),
		)

		return resp, err
	}
}
