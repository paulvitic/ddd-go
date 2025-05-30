package ddd

// HandleEvent defines a function that handles an event
type HandleEvent func(ctx *Context, event Event) error

type EventHandler interface {
	// SubscribedTo returns a map of command types to handler functions
	SubscribedTo() map[string]HandleEvent
}
