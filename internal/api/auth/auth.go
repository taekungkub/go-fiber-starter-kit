package auth

import "time"

// --- DTOs ---

type RegisterDTO struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Telephone string `json:"telephone,omitempty"`
}

type LoginDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
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

type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

// --- Internal Model ---

type UserModel struct {
	ID               string     `db:"id"`
	Email            string     `db:"email"`
	Username         string     `db:"username"`
	Password         string     `db:"password"`
	FirstName        string     `db:"first_name"`
	LastName         string     `db:"last_name"`
	Name             string     `db:"name"`
	Role             string     `db:"role"`
	IsActive         bool       `db:"is_active"`
	RefreshTokenHash *string    `db:"refresh_token_hash"`
	Telephone        *string    `db:"telephone"`
	DeletedAt        *time.Time `db:"deleted_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

// ToResponse converts UserModel to UserResponse
func (u *UserModel) ToResponse() UserResponse {
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
