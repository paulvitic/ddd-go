package eventModelling

import "github.com/paulvitic/ddd-go"

type EventStream func(eventLog ddd.EventLog, command ddd.Command) ddd.Event
