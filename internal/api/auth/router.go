package auth

import (
	"go-fiber-stater-kit/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRouter(app fiber.Router, handler Handler) {
	auth := app.Group("/auth")

	// Public routes
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)

	// Protected routes
	auth.Post("/logout", middleware.JWTMiddleware(), handler.Logout)
}
