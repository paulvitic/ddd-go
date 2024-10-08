package eventModelling

import "github.com/paulvitic/ddd-go"

type EventSource func(log ddd.EventLog, aggregateType, aggregateID string) any
