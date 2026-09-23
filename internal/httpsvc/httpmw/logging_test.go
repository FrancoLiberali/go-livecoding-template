package httpmw_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"interview/internal/httpsvc/httpmw"
)

func TestRequestLogger_LogsRequestAndResponse(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := httpmw.RequestLogger(logger)(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("hi"))
		}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", nil))

	line := buf.String()
	require.Contains(t, line, `"msg":"http_request"`)
	assert.Contains(t, line, `"method":"POST"`)
	assert.Contains(t, line, `"path":"/items"`)
	assert.Contains(t, line, `"status":201`)
	assert.Contains(t, line, `"bytes":2`)
}
