package auth

import (
	"go-fiber-stater-kit/pkg/core"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}

type handler struct {
	Uc UseCase
}

func NewHandler(uc UseCase) Handler {
	return &handler{
		Uc: uc,
	}
}

func (h *handler) Register(c *fiber.Ctx) error {
	var dto RegisterDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	result, err := h.Uc.Register(&dto)
	if err != nil {
		return core.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(core.SuccessResponse{
		Error:   false,
		Message: "Registration successful",
		Data:    result,
	})
}

func (h *handler) Login(c *fiber.Ctx) error {
	var dto LoginDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	result, err := h.Uc.Login(&dto)
	if err != nil {
		return core.SendError(c, fiber.StatusUnauthorized, err.Error())
	}

	return core.SendSuccess(c, "Login successful", result)
}

func (h *handler) RefreshToken(c *fiber.Ctx) error {
	var dto RefreshTokenDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	result, err := h.Uc.RefreshToken(&dto)
	if err != nil {
		return core.SendError(c, fiber.StatusUnauthorized, err.Error())
	}

	return core.SendSuccess(c, "Token refreshed successfully", result)
}

func (h *handler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return core.SendError(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	token, ok := c.Locals("token").(string)
	if !ok {
		return core.SendError(c, fiber.StatusUnauthorized, "Token not found")
	}

	if err := h.Uc.Logout(userID, token); err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, "Failed to logout")
	}

	return core.SendSuccess(c, "Logout successful", nil)
}
