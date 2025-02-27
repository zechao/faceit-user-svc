package tracing_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/zechao/faceit-user-svc/tracing"
)

func TestContextWithTracingID(t *testing.T) {
	ctx := context.Background()
	ctxWithTracingID := tracing.ContextWithTracingID(ctx, uuid.NewString())

	tracingID, ok := tracing.FromContext(ctxWithTracingID)
	if !ok {
		t.Fatalf("expected tracing ID to be present in context")
	}

	if tracingID == "" {
		t.Fatalf("expected tracing ID to be non-empty")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	_, ok := tracing.FromContext(ctx)
	if ok {
		t.Fatalf("expected no tracing ID in context")
	}

	ctxWithTracingID := tracing.ContextWithTracingID(ctx, uuid.NewString())
	tracingID, ok := tracing.FromContext(ctxWithTracingID)
	if !ok {
		t.Fatalf("expected tracing ID to be present in context")
	}

	if tracingID == "" {
		t.Fatalf("expected tracing ID to be non-empty")
	}
}

func TestTracingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets new tracing ID if not provided", func(t *testing.T) {
		router := gin.New()
		router.Use(tracing.TracingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			traceIDFromContext, ok := tracing.FromContext(c.Request.Context())
			assert.True(t, ok)
			assert.NotEmpty(t, traceIDFromContext)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get(tracing.TracingIDKey))
	})

	t.Run("uses provided tracing ID", func(t *testing.T) {
		router := gin.New()
		providedTracingID := uuid.NewString()
		router.Use(tracing.TracingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			traceIDFromContext, ok := tracing.FromContext(c.Request.Context())
			assert.True(t, ok)
			assert.Equal(t, providedTracingID, traceIDFromContext)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(tracing.TracingIDKey, providedTracingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, providedTracingID, w.Header().Get(tracing.TracingIDKey))
	})
}
