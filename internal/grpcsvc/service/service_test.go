package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/repository/mocks"
	"interview/internal/grpcsvc/service"
)

func TestService_Get(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	want := domain.Item{ID: "1", Name: "widget"}
	repo.EXPECT().GetByID(ctx, "1").Return(want, nil)

	got, err := service.New(repo).Get(ctx, "1")

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	repo.EXPECT().
		Save(ctx, mock.MatchedBy(func(item domain.Item) bool {
			return item.Name == "widget" && item.ID != ""
		})).
		Return(nil)

	got, err := service.New(repo).Create(ctx, "widget")

	require.NoError(t, err)
	assert.Equal(t, "widget", got.Name)
	assert.NotEmpty(t, got.ID)
}

func TestService_Create_SaveFails(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockRepository(t)
	repo.EXPECT().Save(ctx, mock.Anything).Return(errors.New("db down"))

	_, err := service.New(repo).Create(ctx, "widget")

	require.Error(t, err)
}
