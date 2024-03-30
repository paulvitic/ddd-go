package go_ddd

// View is a read-only projection of the domain state.
// It is only mutated by particular domain events that it subscribes to.
type View interface {
	SubscribedTo() []string
	MutateWhen(Event) error
}

type view struct {
	subscribedToEvents []string
}

func NewView(subscribedToEvents ...interface{}) View {
	var eventTypes []string
	for _, event := range subscribedToEvents {
		eventTypes = append(eventTypes, EventType(event))
	}
	return &view{
		subscribedToEvents: eventTypes,
	}
}

func (p *view) SubscribedTo() []string {
	return p.subscribedToEvents
}

func (p *view) MutateWhen(e Event) error {
	panic("implement me")
}
