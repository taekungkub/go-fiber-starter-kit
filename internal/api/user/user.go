package user

import "time"

// --- UserRole ---

const (
	RoleAdmin    = "ADMIN"
	RoleStaff    = "STAFF"
	RoleCustomer = "CUSTOMER"
)

// --- User Model ---

type User struct {
	ID               string     `db:"id" json:"id"`
	Email            string     `db:"email" json:"email"`
	Username         string     `db:"username" json:"username"`
	Password         string     `db:"password" json:"-"`
	FirstName        string     `db:"first_name" json:"firstName"`
	LastName         string     `db:"last_name" json:"lastName"`
	Name             string     `db:"name" json:"name"`
	Role             string     `db:"role" json:"role"`
	IsActive         bool       `db:"is_active" json:"isActive"`
	RefreshTokenHash *string    `db:"refresh_token_hash" json:"-"`
	Telephone        *string    `db:"telephone" json:"telephone,omitempty"`
	DeletedAt        *time.Time `db:"deleted_at" json:"-"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updatedAt"`
}

// --- DTOs ---

type CreateUserDTO struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Role      string `json:"role" validate:"required,oneof=ADMIN STAFF CUSTOMER"`
	Telephone string `json:"telephone,omitempty"`
}

type UpdateUserDTO struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Username  *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Role      *string `json:"role,omitempty" validate:"omitempty,oneof=ADMIN STAFF CUSTOMER"`
	IsActive  *bool   `json:"isActive,omitempty"`
	Telephone *string `json:"telephone,omitempty"`
}

// --- Response ---

type UserResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Name      string     `json:"name"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"isActive"`
	Telephone *string    `json:"telephone,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Name:      u.Name,
		Role:      u.Role,
		IsActive:  u.IsActive,
		Telephone: u.Telephone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
