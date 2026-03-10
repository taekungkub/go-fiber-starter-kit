package middleware

import (
	"go-fiber-stater-kit/pkg/core"

	"github.com/gofiber/fiber/v2"
)

// RequireRole checks if the authenticated user has one of the allowed roles
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return core.SendError(c, fiber.StatusForbidden, "Access denied: no role found")
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return core.SendError(c, fiber.StatusForbidden, "Access denied: insufficient permissions")
	}
}
