// Package repository provides item storage implementations. The interface it
// satisfies is declared by its consumer (the service package), following Go's
// "define interfaces where they are used" idiom — so this package exports a
// concrete type, not an interface.
package repository

import (
	"context"
	"sync"

	"interview/internal/httpsvc/domain"
)

// InMemory is a goroutine-safe in-memory item store, handy for tests and demos.
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
