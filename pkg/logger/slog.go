package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewLogger(level LogLevel, environment string) Logger {
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().UTC().Format(time.RFC3339))
			}

			if a.Key == slog.SourceKey && slogLevel <= slog.LevelError {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					source.File = runtime.FuncForPC(0).Name()
				}
			}
			return a
		},
		AddSource: environment == "development",
	}

	var handler slog.Handler
	if environment == "development" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &SlogLogger{logger: logger}
}

func (l *SlogLogger) WithContext(ctx context.Context) Logger {
	args := make([]any, 0)

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		args = append(args, "request_id", requestID.(string))
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		args = append(args, "user_id", userID.(string))
	}
	if operation := ctx.Value(OperationKey); operation != nil {
		args = append(args, "operation", operation.(string))
	}
	if component := ctx.Value(ComponentKey); component != nil {
		args = append(args, "component", component.(string))
	}
	if correlationID := ctx.Value(CorrelationKey); correlationID != nil {
		args = append(args, "correlation_id", correlationID.(string))
	}

	if len(args) > 0 {
		return &SlogLogger{logger: l.logger.With(args...)}
	}

	return l
}

func (l *SlogLogger) WithFields(fields map[string]any) Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &SlogLogger{logger: l.logger.With(args...)}
}

func (l *SlogLogger) WithField(key string, value any) Logger {
	return &SlogLogger{logger: l.logger.With(key, value)}
}

func (l *SlogLogger) WithComponent(component string) Logger {
	return &SlogLogger{logger: l.logger.With("component", component)}
}

func (l *SlogLogger) WithOperation(operation string) Logger {
	return &SlogLogger{logger: l.logger.With("operation", operation)}
}

func (l *SlogLogger) WithError(err error) Logger {
	return &SlogLogger{logger: l.logger.With("error", err.Error())}
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SlogLogger) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Debug(msg, args...)
}

func (l *SlogLogger) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Info(msg, args...)
}

func (l *SlogLogger) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Warn(msg, args...)
}

func (l *SlogLogger) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Error(msg, args...)
}

func (l *SlogLogger) GetSlogLogger() *slog.Logger {
	return l.logger
}

func (l *SlogLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Error(msg, fields...)
	os.Exit(1)
}

func (l *SlogLogger) LogAttrs(ctx context.Context, level slog.Level, msg string, attr ...slog.Attr) {
	l.logger.LogAttrs(ctx, level, msg, attr...)
}
