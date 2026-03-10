package core

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Sorting struct {
	Sort  string // field name (db column)
	Order string // ASC or DESC
}

// SortingRequest parses sort and order query params with validation
// Example: ?sort=created_at&order=asc
func SortingRequest(c *fiber.Ctx, defaultSort string, defaultOrder string) Sorting {
	sort := strings.TrimSpace(c.Query("sort"))
	order := strings.ToUpper(strings.TrimSpace(c.Query("order")))

	// Validate sort field against whitelist

	// Validate order
	if order != "ASC" && order != "DESC" {
		order = defaultOrder
	}

	return Sorting{
		Sort:  sort,
		Order: order,
	}
}
