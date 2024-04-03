package eventModelling

import "github.com/paulvitic/ddd-go"

type EventSource func(log go_ddd.EventLog, aggregateType, aggregateID string) any
