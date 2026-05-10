package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TraceMiddleware creates a new middleware for OpenTelemetry tracing.
func TraceMiddleware() fiber.Handler {
	tracer := otel.Tracer("fiber-server")
	propagator := otel.GetTextMapPropagator()

	return func(c fiber.Ctx) error {
		// Extract propagation context from headers
		ctx := propagator.Extract(c.Context(), propagation.HeaderCarrier(c.GetReqHeaders()))

		// Start a new span
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Path())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.host", c.Hostname()),
				attribute.String("http.route", c.Route().Path),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Inject the new context into the Fiber Ctx
		c.SetContext(ctx)

		// Process the request
		err := c.Next()

		// Record response details
		status := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", status))
		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}

		return err
	}
}
