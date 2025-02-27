package event

import "context"

// EventHandler is an interface that defines the method to send an event to the event bus.
type EventHandler interface {
	SendEvent(ctx context.Context, eventType string, payload any) error
}

// Event represents an event that will be sent to the event bus.
type Event struct {
	TraceID   string `json:"trace_id"`
	EventType string `json:"event_type"`
	Timestamp int64  `json:"timestamp"`
	Payload   any    `json:"payload"`
}
