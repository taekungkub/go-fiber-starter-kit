package user

import (
	"go-fiber-stater-kit/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func UserRouter(app fiber.Router, handler Handler, redisClient *redis.Client) {
	users := app.Group("/users", middleware.JWTMiddleware(redisClient))

	users.Get("/", handler.GetAll)
	users.Get("/:id", handler.GetByID)
	users.Post("/", middleware.RequireRole("ADMIN"), handler.Create)
	users.Patch("/:id", handler.Update)
	users.Delete("/:id", middleware.RequireRole("ADMIN"), handler.Delete)
}
