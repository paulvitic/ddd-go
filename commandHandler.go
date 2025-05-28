package ddd

import "context"

// EventHandler defines a function that handles an event
type HandleCommand func(context.Context, Command) error

// CommandHandler defines a handler that can process specific command types
type CommandHandler interface {
	// SubscribedTo returns a map of command types to handler functions
	SubscribedTo() map[string]HandleCommand
}
