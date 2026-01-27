package middleware

import (
	"strconv"
	"time"

	"hysteria2_microservices/api-service/pkg/metrics"

	"github.com/gofiber/fiber/v2"
)

// Metrics middleware for collecting HTTP metrics
func Metrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment active connections
		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		// Process request
		err := c.Next()

		// Record metrics
		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Path()

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())

		return err
	}
}
