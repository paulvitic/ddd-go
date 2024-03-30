package go_ddd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ByName struct {
	Name string
}
type ByNameAndId struct {
	Name string
	ID   int
}
type AQueryResponseItem struct {
	ID   int
	Name string
}

func TestNewQuery(t *testing.T) {
	q := NewPagedQuery(ByName{Name: "value"}, 1, 10)
	assert.Equal(t, "github.com/paulvitic/ddd-go.ByName", q.Type())
	assert.Equal(t, "value", q.Filter().(ByName).Name)
	assert.Equal(t, 10, q.PageSize())
	assert.Equal(t, 1, q.PageIndex())
}

func TestNewQueryResponse(t *testing.T) {
	items := []AQueryResponseItem{{ID: 1, Name: "value"}, {ID: 2, Name: "value"}}
	qr := NewPagedQueryResponse(items, 10, 0, 2)

	assert.Equal(t, "value", qr.Items().([]AQueryResponseItem)[0].Name)
	assert.Equal(t, 2, qr.Items().([]AQueryResponseItem)[1].ID)
	assert.Equal(t, 5, qr.TotalPages())
	assert.Equal(t, 1, qr.PageNumber())
	assert.False(t, qr.HasPrev())
	assert.Equal(t, 0, qr.Prev())
	assert.True(t, qr.HasNext())
	assert.Equal(t, 2, qr.Next())
}

func TestNewQueryResponseWithPrev(t *testing.T) {
	items := []AQueryResponseItem{{ID: 1, Name: "value"}, {ID: 2, Name: "value"}}
	qr := NewPagedQueryResponse(items, 10, 1, 2)

	assert.True(t, qr.HasPrev())
	assert.Equal(t, 1, qr.Prev())
	assert.True(t, qr.HasNext())
	assert.Equal(t, 3, qr.Next())
}
