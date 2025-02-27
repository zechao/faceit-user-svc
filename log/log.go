package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Logger is an interface for logging messages with context and parameters.
type Logger interface {
	Info(ctx context.Context, msg string, params ...any)
	Error(ctx context.Context, msg string, params ...any)
	Debug(ctx context.Context, msg string, params ...any)
	Warn(ctx context.Context, msg string, params ...any)
}

type SlogLogger struct {
	logger *slog.Logger
}

func NewStdoutJSONLogger(level slog.Level) *SlogLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &SlogLogger{
		logger: slog.New(handler),
	}
}

func (l *SlogLogger) Debug(ctx context.Context, msg string, params ...any) {
	l.logger.DebugContext(ctx, msg, params...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, params ...any) {
	l.logger.InfoContext(ctx, msg, params...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, params ...any) {
	l.logger.ErrorContext(ctx, msg, params...)
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, params ...any) {
	l.logger.WarnContext(ctx, msg, params...)
}

func (l *SlogLogger) Fatal(err error) {
	l.logger.Error(fmt.Sprintf("Fatal error: %v", err))
	os.Exit(1)
}
