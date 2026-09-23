// Package service holds the gRPC service's business logic. It depends only on
// the repository.Repository interface, so it is unit-tested with a mock repo.
package service

import (
	"context"

	"github.com/google/uuid"

	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/repository"
)

// Service is the business-logic port consumed by the controller. Mocked by mockery.
type Service interface {
	Get(ctx context.Context, id string) (domain.Item, error)
	Create(ctx context.Context, name string) (domain.Item, error)
}

type service struct {
	repo repository.Repository
}

// New wires the service with its repository dependency.
func New(repo repository.Repository) Service {
	return &service{repo: repo}
}

// Get returns the item with the given ID.
func (s *service) Get(ctx context.Context, id string) (domain.Item, error) {
	return s.repo.GetByID(ctx, id)
}

// Create persists a new item with a generated ID and returns it.
func (s *service) Create(ctx context.Context, name string) (domain.Item, error) {
	item := domain.Item{ID: uuid.NewString(), Name: name}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.Item{}, err
	}

	return item, nil
}
