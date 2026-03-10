package middleware

import (
	"context"
	"go-fiber-stater-kit/pkg/core"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// JWTMiddleware validates JWT access tokens, checks Redis blacklist, and attaches user info to context
func JWTMiddleware(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return core.SendError(c, fiber.StatusUnauthorized, "Missing authorization header")
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return core.SendError(c, fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		tokenString := parts[1]

		// Check if token is blacklisted in Redis
		ctx := context.Background()
		blacklisted, err := redisClient.Get(ctx, "blacklist:"+tokenString).Result()
		if err == nil && blacklisted == "1" {
			return core.SendError(c, fiber.StatusUnauthorized, "Token has been revoked")
		}

		// Validate token
		claims, err := core.ValidateAccessToken(tokenString)
		if err != nil {
			return core.SendError(c, fiber.StatusUnauthorized, "Invalid or expired token")
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("token", tokenString)

		return c.Next()
	}
}
