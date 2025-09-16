package logger

import (
	"context"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type ContextKey string

const (
	RequestIDKey   ContextKey = "request_id"
	UserIDKey      ContextKey = "user_id"
	OperationKey   ContextKey = "operation"
	ComponentKey   ContextKey = "component"
	CorrelationKey ContextKey = "correlation_id"
)

type Logger interface {
	WithContext(ctx context.Context) Logger
	WithFields(fields map[string]any) Logger
	WithField(key string, value any) Logger
	WithComponent(component string) Logger
	WithOperation(operation string) Logger
	WithError(err error) Logger

	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)

	DebugCtx(ctx context.Context, msg string, args ...any)
	InfoCtx(ctx context.Context, msg string, args ...any)
	WarnCtx(ctx context.Context, msg string, args ...any)
	ErrorCtx(ctx context.Context, msg string, args ...any)

	Fatal(msg string, fields ...interface{})
}

type LoggerFactory func(level LogLevel, environment string) Logger

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, OperationKey, operation)
}

func WithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, ComponentKey, component)
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationKey, correlationID)
}
