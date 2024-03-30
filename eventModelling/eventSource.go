package eventModelling

import "github.com/paulvitic/ddd-go"

type EventSource[I go_ddd.ID] func(log go_ddd.EventLog, aggregateType, aggregateID string) any
