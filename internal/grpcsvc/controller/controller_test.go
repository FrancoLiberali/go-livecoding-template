package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"interview/internal/grpcsvc/controller"
	"interview/internal/grpcsvc/domain"
	"interview/internal/grpcsvc/pb"
	"interview/internal/grpcsvc/service/mocks"
)

func TestController_GetItem_OK(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, "1").Return(domain.Item{ID: "1", Name: "widget"}, nil)

	resp, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: "1"})

	require.NoError(t, err)
	assert.Equal(t, "widget", resp.GetItem().GetName())
}

func TestController_GetItem_NotFound(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, "missing").Return(domain.Item{}, domain.ErrNotFound)

	_, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: "missing"})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestController_CreateItem_OK(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Create(mock.Anything, "widget").Return(domain.Item{ID: "1", Name: "widget"}, nil)

	resp, err := controller.New(svc).CreateItem(context.Background(), &pb.CreateItemRequest{Name: "widget"})

	require.NoError(t, err)
	assert.Equal(t, "1", resp.GetItem().GetId())
}

func TestController_CreateItem_MissingName(t *testing.T) {
	svc := mocks.NewMockService(t)

	_, err := controller.New(svc).CreateItem(context.Background(), &pb.CreateItemRequest{Name: ""})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "name is required", status.Convert(err).Message())
}

func TestController_GetItem_Timeout(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, "slow").RunAndReturn(
		func(ctx context.Context, _ string) (domain.Item, error) {
			<-ctx.Done()

			return domain.Item{}, ctx.Err()
		})

	ctrl := controller.NewWithTimeouts(svc, controller.Timeouts{GetItem: 20 * time.Millisecond})
	_, err := ctrl.GetItem(context.Background(), &pb.GetItemRequest{Id: "slow"})

	require.Error(t, err)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}
