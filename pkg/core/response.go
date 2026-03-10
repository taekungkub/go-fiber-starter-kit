package core

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   bool              `json:"error"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) []ValidationError {
	var errors []ValidationError
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = err.Field()
			element.Message = getErrorMsg(err)
			errors = append(errors, element)
		}
	}
	return errors
}

// getErrorMsg returns a human-readable error message for validation errors
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// SendValidationError sends a validation error response
func SendValidationError(c *fiber.Ctx, errors []ValidationError) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error:   true,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// SendError sends a generic error response
func SendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Error:   true,
		Message: message,
	})
}

// SendSuccess sends a success response
func SendSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(SuccessResponse{
		Error:   false,
		Message: message,
		Data:    data,
	})
}

type Response struct {
	Code    string      `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message interface{} `json:"message"`
	Cause   interface{} `json:"cause,omitempty"`
}

func UnauthorizedAuth(c *fiber.Ctx, data interface{}) error {
	return c.Status(http.StatusUnauthorized).JSON(&Response{
		Code:    "401",
		Message: data,
	})
}
