package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func TenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantSlug := c.Get("X-Tenant-Slug")
		if tenantSlug == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-Slug header is required",
			})
		}
		c.Locals("tenant_slug", tenantSlug)
		return c.Next()
	}
}

func AuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header with Bearer token is required",
			})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		c.Locals("token", token)
		return c.Next()
	}
}

func RequestLoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		return err
	}
}
