package role

import "time"

// --- UserRole ---

const (
	RoleAdmin    = "ADMIN"
	RoleStaff    = "STAFF"
	RoleCustomer = "CUSTOMER"
)

// --- Role Model ---

type Role struct {
	Id          int       `db:"id"               json:"id"`
	RoleName    string    `db:"role_name"        json:"role_name"       validate:"required"`
	Description *string   `db:"description"      json:"description"`
	Status      *string   `db:"status"           json:"status"          validate:"required,oneof=A I"`
	CreatedAt   time.Time `db:"createdAt"        json:"created_at"`
	UpdatedAt   time.Time `db:"updatedAt"        json:"updated_at"`
}

// --- DTOs ---

type CreateRoleDTO struct {
	RoleName    string  `json:"roleName" validate:"required"`
	Description *string `json:"description"`
	Status      *string `json:"status" validate:"required,oneof=A I"`
}

type UpdateRoleDTO struct {
	RoleName    *string `json:"roleName"`
	Description *string `json:"description"`
	Status      *string `json:"status" validate:"omitempty,oneof=A I"`
}

// --- Response ---

type RoleResponse struct {
	ID          string `json:"id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (u *Role) ToResponse() RoleResponse {
	return RoleResponse{
		ID:          string(u.Id),
		RoleName:    u.RoleName,
		Description: *u.Description,
		Status:      *u.Status,
	}
}
