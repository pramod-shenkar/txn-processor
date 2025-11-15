package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const CorrelationIDKey = "X-Correlation-ID"

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {

		corrID := c.Get(CorrelationIDKey)
		if corrID == "" {
			corrID = uuid.NewString()
			c.Set(CorrelationIDKey, corrID)
		}

		// Attach to context for downstream logging
		ctx := context.WithValue(c.Context(), CorrelationIDKey, corrID)
		c.SetUserContext(ctx)

		start := time.Now()
		slog.InfoContext(ctx, "Incoming request", "correlation_id", corrID, "method", c.Method(), "path", c.OriginalURL(), "ip", c.IP())

		// Process request
		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		slog.InfoContext(ctx, "Outgoing response", "correlation_id", corrID, "method", c.Method(), "path", c.OriginalURL(), "status", status, "latency_ms", latency.Milliseconds(), "error", err)

		return err
	}
}
