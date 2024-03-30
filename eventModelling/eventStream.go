package eventModelling

import "github.com/paulvitic/ddd-go"

type EventStream[I go_ddd.ID] func(eventLog go_ddd.EventLog, command go_ddd.Command) go_ddd.Event
