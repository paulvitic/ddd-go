package go_ddd

import (
	"context"
	"testing"
)

func TestQueryBus(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping test in short mode.")
	//}
	var err error

	bus := NewQueryBus()

	handler := func(context.Context, Query) (QueryResponse, error) {
		items := []AQueryResponseItem{{ID: 1, Name: "value"}, {ID: 2, Name: "value"}}
		return NewPagedQueryResponse(items, 10, 1, 2), nil
	}

	service := NewQueryService(handler, ByName{}, ByNameAndId{})

	if err = bus.RegisterService(service); err != nil {
		t.Errorf("RegisterService() = %v, want %v", err, nil)
	}

	res, err := bus.Dispatch(context.Background(), NewQuery(ByName{Name: "value"}))
	if err != nil {
		t.Errorf("Dispatch() = %v, want %v", err, nil)
	}
	if res.Items().([]AQueryResponseItem)[0].ID != 1 {
		t.Errorf("Dispatch() = %v, want %v", res, false)
	}
}
