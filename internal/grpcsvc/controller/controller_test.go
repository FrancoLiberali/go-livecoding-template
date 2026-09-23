package controller_test

import (
	"context"
	"testing"

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

const validID = "123e4567-e89b-12d3-a456-426614174000"

func TestController_GetItem_OK(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{ID: validID, Name: "widget"}, nil)

	resp, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: validID})

	require.NoError(t, err)
	assert.Equal(t, "widget", resp.GetItem().GetName())
}

func TestController_GetItem_NotFound(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, domain.ErrNotFound)

	_, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: validID})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestController_GetItem_InvalidID(t *testing.T) {
	svc := mocks.NewMockService(t)

	_, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: "not-a-uuid"})

	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "id must be a valid UUID", status.Convert(err).Message())
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

// When the deadline derived by the RPC is honored by the service, it returns
// context.DeadlineExceeded, which the RPC surfaces as codes.DeadlineExceeded.
func TestController_GetItem_Timeout(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, context.DeadlineExceeded)

	_, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: validID})

	require.Error(t, err)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
}

// When the client cancels, the service returns context.Canceled, which the RPC
// surfaces as codes.Canceled.
func TestController_GetItem_ClientCanceled(t *testing.T) {
	svc := mocks.NewMockService(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, context.Canceled)

	_, err := controller.New(svc).GetItem(context.Background(), &pb.GetItemRequest{Id: validID})

	require.Error(t, err)
	assert.Equal(t, codes.Canceled, status.Code(err))
}
