// Package service holds the HTTP service's business logic. It declares the
// Repository interface it needs here (consumer-side), so it depends on no
// concrete storage package — the interface is defined at the point of use.
package service

import (
	"context"

	"github.com/google/uuid"

	"interview/internal/httpsvc/domain"
)

// Repository is the storage behavior the service needs. The concrete
// implementation lives elsewhere (e.g. the repository package); declaring the
// interface here lets it be mocked for unit tests.
type Repository interface {
	GetByID(ctx context.Context, id string) (domain.Item, error)
	Save(ctx context.Context, item domain.Item) error
}

// Service implements the item business logic over a Repository.
type Service struct {
	repo Repository
}

// New wires the service with its repository dependency.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Get returns the item with the given ID.
func (s *Service) Get(ctx context.Context, id string) (domain.Item, error) {
	return s.repo.GetByID(ctx, id)
}

// Create persists a new item with a generated ID and returns it.
func (s *Service) Create(ctx context.Context, name string) (domain.Item, error) {
	item := domain.Item{ID: uuid.NewString(), Name: name}
	if err := s.repo.Save(ctx, item); err != nil {
		return domain.Item{}, err
	}

	return item, nil
}
