// Package controller is the gRPC transport layer. It implements the generated
// ItemServiceServer and depends on the service.Service interface, so its RPCs
// are unit-tested with a mock service. Requests are validated with
// go-playground/validator. Each RPC derives a deadline with context.WithTimeout
// (the gRPC analog of the HTTP timeout middleware); when the service honors it
// and returns context.DeadlineExceeded, the RPC surfaces codes.DeadlineExceeded.
package controller

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/pb"
	"interview/internal/grpcsvc/service"
)

// Per-RPC timeouts.
const (
	getItemTimeout    = 2 * time.Second
	createItemTimeout = 3 * time.Second
)

// Controller adapts gRPC requests to the item service.
type Controller struct {
	pb.UnimplementedItemServiceServer

	svc      service.Service
	validate *validator.Validate
}

// New builds a controller over the given service.
func New(svc service.Service) *Controller {
	return &Controller{svc: svc, validate: newValidator()}
}

// createItemInput carries validation tags for the CreateItem request fields.
type createItemInput struct {
	Name string `json:"name" validate:"required,max=100"`
}

// getItemInput carries validation tags for the GetItem request fields.
type getItemInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// GetItem returns an item by ID.
func (c *Controller) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	if err := c.validate.Struct(getItemInput{ID: req.GetId()}); err != nil {
		return nil, status.Error(codes.InvalidArgument, validationMessage(err))
	}

	ctx, cancel := context.WithTimeout(ctx, getItemTimeout)
	defer cancel()

	item, err := c.svc.Get(ctx, req.GetId())
	if err != nil {
		// This RPC owns how its domain errors map to status codes.
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return nil, status.Error(codes.NotFound, "item not found")
		default:
			return nil, transportError(ctx, err)
		}
	}

	return &pb.GetItemResponse{Item: toProto(item)}, nil
}

// CreateItem validates and creates an item.
func (c *Controller) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	if err := c.validate.Struct(createItemInput{Name: req.GetName()}); err != nil {
		return nil, status.Error(codes.InvalidArgument, validationMessage(err))
	}

	ctx, cancel := context.WithTimeout(ctx, createItemTimeout)
	defer cancel()

	item, err := c.svc.Create(ctx, req.GetName())
	if err != nil {
		// No domain errors are expected from Create; only transport-level ones.
		return nil, transportError(ctx, err)
	}

	return &pb.CreateItemResponse{Item: toProto(item)}, nil
}

// transportError maps non-domain (transport/infrastructure) failures. Domain
// errors are mapped per-RPC by each handler; this covers the context-driven
// cases (client cancellation, deadline) and the catch-all internal error.
// Unexpected (internal) errors are logged in full before the sanitized status
// is returned, so the real cause survives while the client sees a generic one.
func transportError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled by client")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timed out")
	default:
		slog.ErrorContext(ctx, "internal error handling request", "err", err)

		return status.Error(codes.Internal, "internal error")
	}
}

func toProto(item domain.Item) *pb.Item {
	return &pb.Item{Id: item.ID, Name: item.Name}
}
