package event

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zechao/faceit-user-svc/tracing"
)

var natsURL string

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "nats:latest",
		ExposedPorts: []string{"4222/tcp"},
		WaitingFor:   wait.ForListeningPort("4222/tcp"),
	}

	natsC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	defer natsC.Terminate(ctx)

	host, err := natsC.Host(ctx)
	if err != nil {
		panic(err)
	}

	port, err := natsC.MappedPort(ctx, "4222")
	if err != nil {
		panic(err)
	}

	natsURL = "nats://" + host + ":" + port.Port()

	m.Run()
}

func TestSendEvent(t *testing.T) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	handler := NewNatsEventHandler(nc, "test-topic")

	t.Run("successful event publish with traceID from context", func(t *testing.T) {
		traceID := uuid.NewString()
		ctx := tracing.ContextWithTracingID(context.Background(), traceID)
		eventType := "test-event"
		payload := map[string]string{"key": "value"}

		event := Event{
			TraceID:   traceID,
			EventType: eventType,
			Timestamp: time.Now().Unix(),
			Payload:   payload,
		}

		eventBytes, err := json.Marshal(event)
		assert.NoError(t, err)

		sub, err := nc.SubscribeSync("test-topic")
		if err != nil {
			t.Fatalf("Failed to subscribe to topic: %v", err)
		}

		err = handler.SendEvent(ctx, eventType, payload)
		assert.NoError(t, err)

		msg, err := sub.NextMsg(5 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, eventBytes, msg.Data)
	})

	t.Run("successful event publish with new traceID", func(t *testing.T) {
		ctx := context.Background()
		eventType := "test-event"
		payload := "data"

		sub, err := nc.SubscribeSync("test-topic")
		if err != nil {
			t.Fatalf("Failed to subscribe to topic: %v", err)
		}

		err = handler.SendEvent(ctx, eventType, payload)
		assert.NoError(t, err)

		msg, err := sub.NextMsg(5 * time.Second)
		assert.NoError(t, err)

		var event Event
		json.Unmarshal(msg.Data, &event)
		assert.NotEmpty(t, event.TraceID)
		assert.Equal(t, eventType, event.EventType)
		assert.Equal(t, payload, event.Payload)
	})

	t.Run("failed to marshal event", func(t *testing.T) {
		ctx := context.Background()
		eventType := "test-event"
		payload := make(chan int) // payload that cannot be marshaled to JSON

		err := handler.SendEvent(ctx, eventType, payload)
		assert.Error(t, err)
	})

	t.Run("failed to publish event", func(t *testing.T) {
		traceID := uuid.NewString()
		ctx := tracing.ContextWithTracingID(context.Background(), traceID)
		eventType := "test-event"
		payload := "data"

		event := Event{
			TraceID:   traceID,
			EventType: eventType,
			Timestamp: time.Now().Unix(),
			Payload:   payload,
		}

		eventBytes, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NotNil(t, eventBytes)
		sub, err := nc.SubscribeSync("test-topic")
		assert.NoError(t, err)
		assert.NotNil(t, sub)

		// Simulate a failure by closing the connection
		nc.Close()

		err = handler.SendEvent(ctx, eventType, payload)
		assert.Error(t, err)
	})
}
func TestNewNatConnection(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		nc, err := NewNatConnection(natsURL)
		assert.NoError(t, err)
		assert.NotNil(t, nc)
		defer nc.Close()
	})

	t.Run("failed connection with invalid URL", func(t *testing.T) {
		_, err := NewNatConnection("invalid-url")
		assert.Error(t, err)
	})
}
