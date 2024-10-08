package ddd

type EventPublisher interface {
	Publish(event Event) error
	Queue() interface{}
	Middleware() MiddlewareFunc
	Close()
}
