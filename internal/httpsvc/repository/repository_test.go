package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"interview/internal/httpsvc/domain"
	"interview/internal/httpsvc/repository"
)

func TestInMemory_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewInMemory()
	want := domain.Item{ID: "1", Name: "widget"}

	require.NoError(t, repo.Save(ctx, want))

	got, err := repo.GetByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestInMemory_GetByID_NotFound(t *testing.T) {
	repo := repository.NewInMemory()

	_, err := repo.GetByID(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
