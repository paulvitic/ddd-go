package http

import (
	ddd "github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockCommandTranslator(r *http.Request) (ddd.Command, error) {
	return nil, nil
}

func mockQueryTranslator(r *http.Request) (ddd.Query, error) {
	return nil, nil
}

func TestEndpoint(t *testing.T) {
	e := NewEndpoint("/test")
	assert.Equal(t, "/test", e.Path())

	e.WithCommandTranslator(mockCommandTranslator)
	assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodDelete}, e.Methods())

	e.WithQueryTranslator(mockQueryTranslator)
	assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodGet}, e.Methods())

	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := e.Handler()
	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
