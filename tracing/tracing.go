// provides a middleware for setting a tracing ID in the request context.
// It also provides a function to retrieve the tracing ID from the context.
// The middleware sets a tracing ID if not provided in the request header.
// It can also define interceptors for gRPC services.
package tracing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TracingIDKey is the key that should be sent in the request header to set the tracing ID.
	TracingIDKey = "X-Trace-Id"
)

var tracingIDKey = &struct{}{}

// ContextWithTracingID returns a new context with a tracing ID.
func ContextWithTracingID(ctx context.Context, tracingID string) context.Context {
	return context.WithValue(ctx, tracingIDKey, tracingID)
}

// FromContext retrieves the tracing ID from the context.
func FromContext(ctx context.Context) (string, bool) {
	tracingID, ok := ctx.Value(tracingIDKey).(string)
	return tracingID, ok
}

// TracingMiddleware is a Gin middleware that sets a tracing ID if not provided.
// The tracing ID is used to trace a request across services.
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tracingID := c.GetHeader(TracingIDKey)
		if tracingID == "" {
			tracingID = uuid.NewString()

		}
		c.Set(TracingIDKey, tracingID)
		c.Header(TracingIDKey, tracingID)
		c.Request = c.Request.WithContext(ContextWithTracingID(c.Request.Context(), tracingID))
		c.Next()
	}
}
