package http

import (
	ddd "github.com/paulvitic/ddd-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockQueryTranslator(r *http.Request) (ddd.Query, error) {
	return nil, nil
}

func TestQueryEndpoint(t *testing.T) {
	e := NewQueryEndpoint("/test", []string{http.MethodGet}, mockQueryTranslator)
	assert.Equal(t, "/test", e.Path())

	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := e.Handler()
	handler(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
