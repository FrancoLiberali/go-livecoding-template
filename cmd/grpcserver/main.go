// Command grpcserver wires the gRPC service (repository -> service ->
// controller) and serves it over grpc.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"interview/internal/grpcsvc/controller"
	"interview/internal/grpcsvc/grpcmw"
	"interview/internal/grpcsvc/pb"
	"interview/internal/grpcsvc/repository"
	"interview/internal/grpcsvc/service"
)

const addr = ":9090"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("grpc server failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	repo := repository.NewInMemory()
	svc := service.New(repo)
	ctrl := controller.New(svc)

	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	// Logger is outermost so it records every RPC; recovery is closest to the
	// handler so a handler panic is turned into an Internal error the logger
	// still sees (rather than crashing the server).
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpcmw.UnaryLogger(slog.Default()),
		grpcmw.UnaryRecovery(slog.Default()),
	))
	pb.RegisterItemServiceServer(server, ctrl)

	// Standard gRPC health service (grpc.health.v1.Health) for probes.
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.ItemService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)

	reflection.Register(server)

	go func() {
		<-ctx.Done()
		slog.Info("shutting down grpc server")
		server.GracefulStop()
	}()

	slog.Info("grpc server listening", "addr", addr)

	// Serve returns nil once GracefulStop completes.
	return server.Serve(listener)
}
