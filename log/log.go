package log

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zechao/faceit-user-svc/config"
	"github.com/zechao/faceit-user-svc/tracing"
)

var (
	Info    func(ctx context.Context, msg string, params ...any)
	Warn    func(ctx context.Context, msg string, params ...any)
	Error   func(ctx context.Context, msg string, params ...any)
	Fatal   func(err error)
	Fatalf  func(format string, v ...any)
	Printf  func(format string, v ...any)
	Println func(v ...any)

	once sync.Once
)

func init() {
	// Initialize the logger only once
	once.Do(func() {
		logger := NewStdoutJSONLogger(slog.Level(config.ENVs.LogLevel))

		Info = logger.Info
		Warn = logger.Warn
		Error = logger.Error
		Fatal = logger.Fatal
		Fatalf = log.Fatalf
		Printf = log.Printf
		Println = log.Println
	})

}

// SlogLogger is a logger that uses the slog package.
type SlogLogger struct {
	logger *slog.Logger
}

// NewStdoutJSONLogger creates a new logger that writes JSON logs to stdout.
func NewStdoutJSONLogger(level slog.Level) *SlogLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

// Info logs an info message with the provided parameters. It will also log the trace ID if present in the context.
func (l *SlogLogger) Info(ctx context.Context, msg string, params ...any) {
	traceID, ok := tracing.FromContext(ctx)
	if ok {
		params = append(params, "trace_id", traceID)
	}
	l.logger.InfoContext(ctx, msg, params...)
}

// Error logs an error message with the provided parameters. It will also log the trace ID if present in the context.
func (l *SlogLogger) Error(ctx context.Context, msg string, params ...any) {
	traceID, ok := tracing.FromContext(ctx)
	if ok {
		params = append(params, "trace_id", traceID)
	}
	l.logger.ErrorContext(ctx, msg, params...)
}

// Warn logs a warning message with the provided parameters. It will also log the trace ID if present in the context.
func (l *SlogLogger) Warn(ctx context.Context, msg string, params ...any) {
	traceID, ok := tracing.FromContext(ctx)
	if ok {
		params = append(params, "trace_id", traceID)
	}
	l.logger.WarnContext(ctx, msg, params...)
}

// Fatal logs a fatal error message and exits the program with exit code 1.
func (l *SlogLogger) Fatal(err error) {
	l.logger.Error(fmt.Sprintf("Fatal error: %v", err))
	os.Exit(1)
}

// GinLoggerMiddleware logs the request and response, including the status code,
// duration, client IP, method, path, and response size. but never logs the request body or response body.
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()
		// Log request details
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		responseSize := c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}

		Info(c.Request.Context(), "HTTP REQUEST",
			"status_code", statusCode,
			"duration", duration,
			"client_ip", clientIP,
			"method", method,
			"path", path,
			"raw", raw,
			"response_size", responseSize,
		)

	}
}
