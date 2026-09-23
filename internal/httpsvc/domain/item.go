// Package domain holds the HTTP service's core entities and errors,
// independent of any transport or storage concern.
package domain

import "errors"

// ErrNotFound is returned when an item does not exist.
var ErrNotFound = errors.New("item not found")

// Item is the core entity of the HTTP service skeleton.
type Item struct {
	ID   string
	Name string
}
