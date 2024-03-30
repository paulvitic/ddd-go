package go_ddd

// Policy receives a domain event and returns a command.
// It is used to implement domain logic that can be defined as
// when this happens then do that.
type Policy interface {
	SubscribedTo() []string
	When(event Event) (Command, error)
}

type policy struct {
	subscribedToEvents []string
}

func NewPolicy(subscribedToEvents ...interface{}) Policy {
	var eventTypes []string
	for _, event := range subscribedToEvents {
		eventTypes = append(eventTypes, EventType(event))
	}
	return &policy{
		subscribedToEvents: eventTypes,
	}
}

func (p *policy) When(_ Event) (Command, error) {
	panic("implement me")
}

func (p *policy) SubscribedTo() []string {
	return p.subscribedToEvents
}
