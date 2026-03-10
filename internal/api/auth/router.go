package auth

import (
	"go-fiber-stater-kit/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func AuthRouter(app fiber.Router, handler Handler, redisClient *redis.Client) {
	auth := app.Group("/auth")

	// Public routes
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)

	// Protected routes
	auth.Post("/logout", middleware.JWTMiddleware(redisClient), handler.Logout)
}
