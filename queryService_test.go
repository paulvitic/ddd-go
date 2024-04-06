package ddd

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryService(t *testing.T) {
	executor := func(context.Context, Query) (QueryResponse, error) {
		return NewQueryResponse("response"), nil
	}
	service := NewQueryService(executor, ByName{}, ByNameAndId{})

	assert.Equal(t, service.SubscribedTo()[0], "github.com/paulvitic/ddd-go.ByName")

	res, err := executor(context.Background(), NewQuery(ByName{Name: "value"}))
	assert.NoError(t, err)
	assert.Equal(t, res, NewQueryResponse("response"))
}
