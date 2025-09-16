package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zainokta/item-sync/pkg/logger"
)

func LogMiddleware(logger logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Log request details
			req := c.Request()
			res := c.Response()

			duration := time.Since(start)

			// Prepare log fields
			fields := map[string]interface{}{
				"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
				"method":     req.Method,
				"path":       req.URL.Path,
				"remote_ip":  c.RealIP(),
				"user_agent": req.UserAgent(),
				"status":     res.Status,
				"bytes_in":   req.ContentLength,
				"bytes_out":  res.Size,
				"latency":    duration,
			}

			// Add error if present
			if err != nil {
				fields["error"] = err.Error()
			}

			ctxLogger := logger.WithFields(fields).WithComponent("http_server")

			switch {
			case res.Status >= 500:
				ctxLogger.Error("Request completed")
			case res.Status >= 400:
				ctxLogger.Warn("Request completed")
			default:
				ctxLogger.Info("Request completed")
			}

			return err
		}
	}
}
