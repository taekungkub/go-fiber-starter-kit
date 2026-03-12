package ask

import (
	"github.com/gofiber/fiber/v2"
)

func AskRouter(app fiber.Router, handler Handler) {
	users := app.Group("/ask")

	users.Post("/", handler.Create)
}
