package controller_test

import (
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

func newRouter(t *testing.T) (*mocks.MockService, http.Handler) {
	t.Helper()

	svc := mocks.NewMockService(t)
	router := chi.NewRouter()
	controller.New(svc).RegisterRoutes(router)

	return svc, router
}

func TestController_GetItem_OK(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, "1").Return(domain.Item{ID: "1", Name: "widget"}, nil)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items/1", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"id":"1","name":"widget"}`, rec.Body.String())
}

func TestController_GetItem_NotFound(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Get(mock.Anything, "missing").Return(domain.Item{}, domain.ErrNotFound)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items/missing", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestController_CreateItem_OK(t *testing.T) {
	svc, router := newRouter(t)
	svc.EXPECT().Create(mock.Anything, "widget").Return(domain.Item{ID: "1", Name: "widget"}, nil)

	body := strings.NewReader(`{"name":"widget"}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", body))

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"id":"1","name":"widget"}`, rec.Body.String())
}

func TestController_CreateItem_MissingName(t *testing.T) {
	_, router := newRouter(t)

	body := strings.NewReader(`{"name":""}`)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
