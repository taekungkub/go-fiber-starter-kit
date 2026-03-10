package auth

import (
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateUser(dto *RegisterDTO, hashedPassword string) (*UserModel, error)
	FindByEmail(email string) (*UserModel, error)
	FindByUsername(username string) (*UserModel, error)
	UpdateRefreshTokenHash(userID string, hash string) error
	GetRefreshTokenHash(userID string) (*string, error)
	ClearRefreshTokenHash(userID string) error
}

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{DB: db}
}

func (r *repository) CreateUser(dto *RegisterDTO, hashedPassword string) (*UserModel, error) {
	var user UserModel
	query := `
		INSERT INTO users (email, username, password, first_name, last_name, telephone, role)
		VALUES ($1, $2, $3, $4, $5, $6, 'CUSTOMER')
		RETURNING id, email, username, password, first_name, last_name, name, role, is_active, 
		          refresh_token_hash, telephone, deleted_at, created_at, updated_at
	`
	err := r.DB.QueryRowx(query,
		dto.Email, dto.Username, hashedPassword,
		dto.FirstName, dto.LastName, dto.Telephone,
	).StructScan(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByEmail(email string) (*UserModel, error) {
	var user UserModel
	query := `SELECT id, email, username, password, first_name, last_name, name, role, is_active, 
	                 refresh_token_hash, telephone, deleted_at, created_at, updated_at
	          FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByUsername(username string) (*UserModel, error) {
	var user UserModel
	query := `SELECT id, email, username, password, first_name, last_name, name, role, is_active, 
	                 refresh_token_hash, telephone, deleted_at, created_at, updated_at
	          FROM users WHERE username = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateRefreshTokenHash(userID string, hash string) error {
	query := `UPDATE users SET refresh_token_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.DB.Exec(query, hash, userID)
	return err
}

func (r *repository) GetRefreshTokenHash(userID string) (*string, error) {
	var hash *string
	query := `SELECT refresh_token_hash FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&hash, query, userID)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (r *repository) ClearRefreshTokenHash(userID string) error {
	query := `UPDATE users SET refresh_token_hash = NULL, updated_at = NOW() WHERE id = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}
