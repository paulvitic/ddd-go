package ddd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type Event interface {
	AggregateType() string
	AggregateID() ID
	Type() string
	TimeStamp() time.Time
	Payload() any
	ToJsonString() (string, error)
}

type event struct {
	aggregateType string
	aggregateID   ID
	eventType     string
	timeStamp     time.Time
	payload       any
}

func (e *event) AggregateType() string {
	return e.aggregateType
}

func (e *event) AggregateID() ID {
	return e.aggregateID
}

func (e *event) Type() string {
	return e.eventType
}

func (e *event) TimeStamp() time.Time {
	return e.timeStamp
}

func (e *event) Payload() any {
	return e.payload
}

func (e *event) ToJsonString() (string, error) {
	data, err := json.Marshal(map[string]any{
		"aggregate_type": e.aggregateType,
		"aggregate_id":   e.aggregateID.String(),
		"event_type":     e.eventType,
		"time_stamp":     e.timeStamp,
		"payload":        e.payload,
	})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func EventType(eventPayload any) string {
	return reflect.TypeOf(eventPayload).PkgPath() + "." + reflect.TypeOf(eventPayload).Name()
}

func EventFromJsonString(jsonString string) (Event, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil, err
	}

	timeStamp, err := time.Parse(time.RFC3339, data["time_stamp"].(string))
	if err != nil {
		return nil, err
	}

	return &event{
		aggregateType: data["aggregate_type"].(string),
		aggregateID:   NewID(data["aggregate_id"].(string)),
		eventType:     data["event_type"].(string),
		timeStamp:     timeStamp,
		payload:       data["payload"],
	}, nil
}

func MapEventPayload[T any](event Event, payload T) T {
	jsonStr, err := json.Marshal(event.Payload())
	if err != nil {
		fmt.Println(err)
	}
	if err := json.Unmarshal(jsonStr, &payload); err != nil {
		fmt.Println(err)
	}
	return payload
}
