package role

import (
	"go-fiber-stater-kit/pkg/core"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type handler struct {
	Uc UseCase
}

func NewHandler(uc UseCase) Handler {
	return &handler{
		Uc: uc,
	}
}

func (h *handler) GetAll(c *fiber.Ctx) error {
	page, limit := core.PagingRequest(c, 0) // default 0 = no limit (return all)
	sorting := core.SortingRequest(c, "created_at", "DESC")

	result, err := h.Uc.FindAll(page, limit, sorting.Sort, sorting.Order)
	if err != nil {
		return core.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return core.SendSuccess(c, "Users retrieved successfully", result)
}

func (h *handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return core.SendError(c, fiber.StatusBadRequest, "User ID is required")
	}

	result, err := h.Uc.FindByID(id)
	if err != nil {
		return core.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return core.SendSuccess(c, "User retrieved successfully", result)
}

func (h *handler) Create(c *fiber.Ctx) error {
	var dto CreateRoleDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	result, err := h.Uc.Create(&dto)
	if err != nil {
		return core.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(core.SuccessResponse{
		Error:   false,
		Message: "User created successfully",
		Data:    result,
	})
}

func (h *handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return core.SendError(c, fiber.StatusBadRequest, "User ID is required")
	}

	var dto UpdateRoleDTO

	if err := c.BodyParser(&dto); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errors := core.ValidateStruct(&dto); errors != nil {
		return core.SendValidationError(c, errors)
	}

	result, err := h.Uc.Update(id, &dto)
	if err != nil {
		return core.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return core.SendSuccess(c, "User updated successfully", result)
}

func (h *handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return core.SendError(c, fiber.StatusBadRequest, "User ID is required")
	}

	if err := h.Uc.Delete(id); err != nil {
		return core.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return core.SendSuccess(c, "User deleted successfully", nil)
}
