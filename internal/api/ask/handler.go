package ask

import (
	"go-fiber-stater-kit/pkg/core"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Create(c *fiber.Ctx) error
}

type handler struct {
	Uc UseCase
}

func NewHandler(uc UseCase) Handler {
	return &handler{
		Uc: uc,
	}
}

func (h *handler) Create(c *fiber.Ctx) error {
	var dto CreateChatDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	answer, err := h.Uc.Create(&dto)
	if err != nil {
		return core.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(core.SuccessResponse{
		Error:   false,
		Message: "Ask successfully",
		Data:    answer,
	})
}
