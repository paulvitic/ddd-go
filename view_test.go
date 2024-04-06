package ddd

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// //////////////////////////////////////
// DomainEvents
// //////////////////////////////////////
type viewNameUpdated struct {
	ID   int
	Name string
}

type viewItemInserted struct {
	ID   int
	Name string
}

// ////////////////////////////////////////
// View
// ////////////////////////////////////////
type testViewItem struct {
	ID   int
	Name string
}
type TestView interface {
	View
	Item(id int) *testViewItem
	AllItems() []*testViewItem

	// For testing purposes
	OnEventCalledTimes() int
}

type abstractTestView struct {
	View
	updateItemName func(id int, name string)
	add            func(item *testViewItem)
	eventCalled    func()
}

func (v *abstractTestView) MutateWhen(event Event) error {
	v.eventCalled()
	switch event.Type() {
	case "github.com/paulvitic/ddd-go.viewItemInserted":
		payload := MapEventPayload(event, viewItemInserted{})
		v.add(&testViewItem{ID: payload.ID, Name: payload.Name})
		return nil
	case "github.com/paulvitic/ddd-go.viewNameUpdated":
		payload := MapEventPayload(event, viewNameUpdated{})
		v.updateItemName(payload.ID, payload.Name)
		return nil
	default:
		return errors.New("unknown event type")
	}
}

// //////////////////////////////////////////////////
// concrete implementation at infrastructure layer
// //////////////////////////////////////////////////
type testInMemoryView struct {
	View
	// In-memory store
	store map[int]*testViewItem
	// For testing purposes
	onEventCalledTimes func() int
}

func NewAViewInMemory() TestView {
	store := make(map[int]*testViewItem)
	onEventCalled := 0

	updateItemName := func(id int, name string) {
		item := store[id]
		item.Name = name
		return
	}

	add := func(item *testViewItem) {
		store[item.ID] = item
	}

	eventCalled := func() {
		onEventCalled++
	}

	onEventCalledTimes := func() int {
		return onEventCalled
	}

	return &testInMemoryView{
		&abstractTestView{
			NewView(viewItemInserted{}, viewNameUpdated{}),
			updateItemName,
			add,
			eventCalled,
		},
		store,
		onEventCalledTimes,
	}
}

// ////////////////////////////////
// Query methods
// ////////////////////////////////
func (v *testInMemoryView) Item(id int) *testViewItem {
	return v.store[id]
}

func (v *testInMemoryView) AllItems() []*testViewItem {
	values := make([]*testViewItem, 0, len(v.store))
	for _, v := range v.store {
		values = append(values, v)
	}
	return values
}

// /////////////////////////////////////
// For testing purposes
// /////////////////////////////////////
func (v *testInMemoryView) OnEventCalledTimes() int {
	return v.onEventCalledTimes()
}

func TestViewBase_SubscribedToEvents(t *testing.T) {
	view := NewAViewInMemory()

	assert.Equal(t, 2, len(view.SubscribedTo()))
	assert.Equal(t, "github.com/paulvitic/ddd-go.viewItemInserted", view.SubscribedTo()[0])
	assert.Equal(t, "github.com/paulvitic/ddd-go.viewNameUpdated", view.SubscribedTo()[1])
}

func TestView_mutation(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping test in short mode.")
	//}
	view := NewAViewInMemory()

	viewItemId := 1

	ep := NewEventProducer().
		RegisterEvent("aggType", NewID("ID123"), viewItemInserted{ID: viewItemId, Name: ""}).
		RegisterEvent("aggType", NewID("ID123"), viewNameUpdated{ID: viewItemId, Name: "value"})

	ev := ep.Events()
	err := view.MutateWhen(ev[0])
	assert.NoError(t, err)
	err = view.MutateWhen(ev[1])
	assert.NoError(t, err)

	assert.Equal(t, 2, view.OnEventCalledTimes())
	assert.Equal(t, "value", view.Item(viewItemId).Name)
}
