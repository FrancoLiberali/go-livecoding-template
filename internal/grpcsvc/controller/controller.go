// Package controller is the gRPC transport layer. It implements the generated
// ItemServiceServer and depends on the service.Service interface, so its RPCs
// are unit-tested with a mock service.
package controller

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/pb"
	"interview/internal/grpcsvc/service"
)

// Controller adapts gRPC requests to the item service.
type Controller struct {
	pb.UnimplementedItemServiceServer

	svc service.Service
}

// New builds a controller over the given service.
func New(svc service.Service) *Controller {
	return &Controller{svc: svc}
}

// GetItem returns an item by ID, mapping a missing item to codes.NotFound.
func (c *Controller) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.GetItemResponse, error) {
	item, err := c.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "item not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.GetItemResponse{Item: toProto(item)}, nil
}

// CreateItem validates and creates an item.
func (c *Controller) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	item, err := c.svc.Create(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.CreateItemResponse{Item: toProto(item)}, nil
}

func toProto(item domain.Item) *pb.Item {
	return &pb.Item{Id: item.ID, Name: item.Name}
}
