package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/zechao/faceit-user-svc/tracing"
)

// Event represents an event that will be sent to the event bus.
type NatsEventHandler struct {
	natsConn *nats.Conn
	topic    string
}

// NewNatConnection is a function type that creates a new nats connection with the provided url.
func NewNatConnection(url string) (*nats.Conn, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return conn, nil
}

// NewNatsEventHandler creates a new NatsEventHandler with the provided nats connection and topic.
func NewNatsEventHandler(natsConn *nats.Conn, topic string) *NatsEventHandler {
	return &NatsEventHandler{
		natsConn: natsConn,
		topic:    topic,
	}
}

// SendEvent sends an event to the event bus with the provided event type and payload.
func (h *NatsEventHandler) SendEvent(ctx context.Context, eventType string, payload any) error {
	traceID, ok := tracing.FromContext(ctx)
	// log if traceID not found in context, but continue to send event with new traceID
	if !ok {
		traceID = uuid.NewString()
		log.Printf("TraceID not found in context, generate new traceID: %s", traceID)
	}

	event := Event{
		TraceID:   traceID,
		EventType: eventType,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	err = h.natsConn.Publish(h.topic, eventBytes)
	if err != nil {
		return err
	}

	return nil
}

// Subscribe simulates another service subscribing to the topic and handling incoming events.
func (h *NatsEventHandler) Subscribe(handler func(event Event)) error {
	_, err := h.natsConn.Subscribe(h.topic, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			return
		}
		handler(event)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", h.topic, err)
	}
	return nil
}
