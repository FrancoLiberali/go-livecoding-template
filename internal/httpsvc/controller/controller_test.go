package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"interview/internal/httpsvc/controller"
	"interview/internal/httpsvc/domain"
	"interview/internal/httpsvc/service/mocks"
)

const validID = "123e4567-e89b-12d3-a456-426614174000"

func newRouter(t *testing.T) (*mocks.MockService, http.Handler) {
	t.Helper()

	svc := mocks.NewMockService(t)
	router := chi.NewRouter()
	controller.New(svc).RegisterRoutes(router)

	return svc, router
}

func do(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func TestController_GetItem_OK(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{ID: validID, Name: "widget"}, nil)

	rec := do(router, httptest.NewRequest(http.MethodGet, "/items/"+validID, nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"id":"`+validID+`","name":"widget"}`, rec.Body.String())
}

func TestController_GetItem_NotFound(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, domain.ErrNotFound)

	rec := do(router, httptest.NewRequest(http.MethodGet, "/items/"+validID, nil))

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"not_found","message":"item not found"}}`, rec.Body.String())
}

func TestController_GetItem_InvalidID(t *testing.T) {
	_, router := newRouter(t)

	rec := do(router, httptest.NewRequest(http.MethodGet, "/items/not-a-uuid", nil))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{
		"error":{
			"code":"validation_error",
			"message":"request validation failed",
			"details":[{"field":"id","message":"id must be a valid UUID"}]
		}
	}`, rec.Body.String())
}

func TestController_CreateItem_OK(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Create(mock.Anything, "widget").Return(domain.Item{ID: "1", Name: "widget"}, nil)

	body := strings.NewReader(`{"name":"widget"}`)
	rec := do(router, httptest.NewRequest(http.MethodPost, "/items", body))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"id":"1","name":"widget"}`, rec.Body.String())
}

func TestController_CreateItem_MissingName(t *testing.T) {
	_, router := newRouter(t)

	body := strings.NewReader(`{"name":""}`)
	rec := do(router, httptest.NewRequest(http.MethodPost, "/items", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{
		"error":{
			"code":"validation_error",
			"message":"request validation failed",
			"details":[{"field":"name","message":"name is required"}]
		}
	}`, rec.Body.String())
}

// When the deadline set by the timeout middleware is honored by the service,
// it returns context.DeadlineExceeded, which the handler surfaces as 504.
func TestController_GetItem_Timeout(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, context.DeadlineExceeded)

	rec := do(router, httptest.NewRequest(http.MethodGet, "/items/"+validID, nil))

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"timeout","message":"request timed out"}}`, rec.Body.String())
}

// When the client cancels before the server responds, the service returns
// context.Canceled, which the handler surfaces as 499 (Client Closed Request).
func TestController_GetItem_ClientCanceled(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, validID).Return(domain.Item{}, context.Canceled)

	rec := do(router, httptest.NewRequest(http.MethodGet, "/items/"+validID, nil))

	require.Equal(t, 499, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"canceled","message":"request canceled by client"}}`, rec.Body.String())
}

func TestController_CreateItem_MalformedJSON(t *testing.T) {
	_, router := newRouter(t)

	body := strings.NewReader(`{"name":`)
	rec := do(router, httptest.NewRequest(http.MethodPost, "/items", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"invalid_request","message":"malformed JSON body"}}`, rec.Body.String())
}
