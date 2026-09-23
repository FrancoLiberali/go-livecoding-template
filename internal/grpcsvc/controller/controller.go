// Package controller is the gRPC transport layer. It implements the generated
// ItemServiceServer and depends on the service.Service interface, so its RPCs
// are unit-tested with a mock service. Requests are validated with
// go-playground/validator, and each RPC runs under its own timeout that yields
// codes.DeadlineExceeded (the gRPC analog of HTTP 504) when exceeded.
package controller

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/pb"
	"interview/internal/grpcsvc/service"
)

const (
	defaultGetItemTimeout    = 2 * time.Second
	defaultCreateItemTimeout = 3 * time.Second
)

// Timeouts holds the per-RPC timeouts.
type Timeouts struct {
	GetItem    time.Duration
	CreateItem time.Duration
}

// DefaultTimeouts returns sensible per-RPC defaults.
func DefaultTimeouts() Timeouts {
	return Timeouts{
		GetItem:    defaultGetItemTimeout,
		CreateItem: defaultCreateItemTimeout,
	}
}

// Controller adapts gRPC requests to the item service.
type Controller struct {
	pb.UnimplementedItemServiceServer

	svc      service.Service
	validate *validator.Validate
	timeouts Timeouts
}

// New builds a controller with the default per-RPC timeouts.
func New(svc service.Service) *Controller {
	return NewWithTimeouts(svc, DefaultTimeouts())
}

// NewWithTimeouts builds a controller with explicit timeouts (used in tests).
func NewWithTimeouts(svc service.Service, timeouts Timeouts) *Controller {
	return &Controller{svc: svc, validate: newValidator(), timeouts: timeouts}
}

// createItemInput carries validation tags for the CreateItem request fields.
type createItemInput struct {
	Name string `json:"name" validate:"required,max=100"`
}

// GetItem returns an item by ID, mapping failures to gRPC status codes.
func (c *Controller) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	item, err := runWithTimeout(ctx, c.timeouts.GetItem,
		func(ctx context.Context) (domain.Item, error) {
			return c.svc.Get(ctx, req.GetId())
		})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.GetItemResponse{Item: toProto(item)}, nil
}

// CreateItem validates and creates an item.
func (c *Controller) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	if err := c.validate.Struct(createItemInput{Name: req.GetName()}); err != nil {
		return nil, status.Error(codes.InvalidArgument, validationMessage(err))
	}

	item, err := runWithTimeout(ctx, c.timeouts.CreateItem,
		func(ctx context.Context) (domain.Item, error) {
			return c.svc.Create(ctx, req.GetName())
		})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &pb.CreateItemResponse{Item: toProto(item)}, nil
}

// toStatusError maps a service-layer error to the right gRPC status.
func toStatusError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "request timed out")
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "item not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func toProto(item domain.Item) *pb.Item {
	return &pb.Item{Id: item.ID, Name: item.Name}
}
