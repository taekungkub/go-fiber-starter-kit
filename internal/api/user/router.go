package user

import (
	"go-fiber-stater-kit/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRouter(app fiber.Router, handler Handler) {
	users := app.Group("/users", middleware.JWTMiddleware())

	users.Get("/", handler.GetAll)
	users.Get("/:id", handler.GetByID)
	users.Post("/", middleware.RequireRole("ADMIN"), handler.Create)
	users.Patch("/:id", handler.Update)
	users.Delete("/:id", middleware.RequireRole("ADMIN"), handler.Delete)
}
