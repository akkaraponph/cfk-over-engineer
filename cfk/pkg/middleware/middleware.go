package middleware

import (
	"cfk/internal/identity/tenant"
	"cfk/pkg/auth"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TenantMiddleware(tenantService *tenant.Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		slug := c.Get("X-Tenant-Slug")
		if slug == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "X-Tenant-Slug header is required",
			})
		}

		t, err := tenantService.GetTenantBySlug(slug).Get()
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "tenant not found",
			})
		}

		if !t.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "tenant is deactivated",
			})
		}

		c.Locals("tenant_id", t.ID)
		c.Locals("tenant_slug", t.Slug)
		c.Locals("tenant_plan", string(t.Plan))
		return c.Next()
	}
}

func FeatureGuard(tenantService *tenant.Service, feature string) fiber.Handler {
	return func(c fiber.Ctx) error {
		tenantID, ok := c.Locals("tenant_id").(string)
		if !ok || tenantID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "tenant context not found",
			})
		}

		if !tenantService.HasFeature(tenantID, feature) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "feature not enabled for tenant: " + feature,
			})
		}

		c.Locals("feature_"+feature, true)
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

		claims, err := auth.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("user_role", claims.Role)
		c.Locals("user_email", claims.Email)
		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("user_role").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied",
			})
		}
		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

func RequestLoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		tenantID := c.Locals("tenant_id")
		userID := c.Locals("user_id")

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		log.Printf("[%s] %s %s %d %s tenant=%v user=%v ip=%s",
			start.Format(time.RFC3339),
			method,
			path,
			status,
			duration.Round(time.Millisecond),
			tenantID,
			userID,
			c.IP(),
		)

		return err
	}
}
