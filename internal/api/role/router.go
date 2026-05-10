package role

import (
	"go-fiber-stater-kit/middleware"

	"github.com/gofiber/fiber/v2"
)

func RoleRouter(app fiber.Router, handler Handler) {
	users := app.Group("/role", middleware.JWTMiddleware())

	users.Get("/", handler.GetAll)
	users.Get("/:id", handler.GetByID)
	users.Post("/", handler.Create)
	users.Patch("/:id", handler.Update)
	users.Delete("/:id", handler.Delete)
}
