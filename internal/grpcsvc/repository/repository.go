// Package repository defines the storage interface for the gRPC service and
// ships a simple in-memory implementation. Depend on the interface, not the
// implementation, so it can be mocked (see ./mocks) and swapped for a real DB.
package repository

import (
	"context"
	"sync"

	"interview/internal/grpcsvc/domain"
)

// Repository is the persistence port for items. Mocked by mockery.
type Repository interface {
	GetByID(ctx context.Context, id string) (domain.Item, error)
	Save(ctx context.Context, item domain.Item) error
}

// InMemory is a goroutine-safe in-memory Repository, handy for tests and demos.
type InMemory struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

// NewInMemory returns an empty in-memory repository.
func NewInMemory() *InMemory {
	return &InMemory{items: make(map[string]domain.Item)}
}

// GetByID returns the item or domain.ErrNotFound.
func (r *InMemory) GetByID(_ context.Context, id string) (domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return domain.Item{}, domain.ErrNotFound
	}

	return item, nil
}

// Save inserts or updates an item by its ID.
func (r *InMemory) Save(_ context.Context, item domain.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[item.ID] = item

	return nil
}
