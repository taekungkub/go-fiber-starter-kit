package user

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindAll(limit, offset int64, sort, order string) ([]User, error)
	Count() (int64, error)
	FindByID(id string) (*User, error)
	Create(dto *CreateUserDTO, hashedPassword string) (*User, error)
	Update(id string, dto *UpdateUserDTO) (*User, error)
	SoftDelete(id string) error
}

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{DB: db}
}

func (r *repository) FindAll(limit, offset int64, sort, order string) ([]User, error) {
	var users []User

	query := fmt.Sprintf(
		`SELECT id, email, username, password, first_name, last_name, name, role, is_active,
		        refresh_token_hash, telephone, deleted_at, created_at, updated_at
		 FROM users WHERE deleted_at IS NULL
		 ORDER BY %s %s`, sort, order,
	)

	var err error
	if limit > 0 {
		query += " LIMIT $1 OFFSET $2"
		err = r.DB.Select(&users, query, limit, offset)
	} else {
		// No limit — return all records
		err = r.DB.Select(&users, query)
	}

	if err != nil {
		return nil, err
	}
	if users == nil {
		users = []User{}
	}

	return users, nil
}

func (r *repository) Count() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := r.DB.Get(&count, query)
	return count, err
}

func (r *repository) FindByID(id string) (*User, error) {
	var user User
	query := `SELECT id, email, username, password, first_name, last_name, name, role, is_active,
	                 refresh_token_hash, telephone, deleted_at, created_at, updated_at
	          FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Create(dto *CreateUserDTO, hashedPassword string) (*User, error) {
	var user User
	query := `INSERT INTO users (email, username, password, first_name, last_name, role, telephone)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING id, email, username, password, first_name, last_name, name, role, is_active,
	                    refresh_token_hash, telephone, deleted_at, created_at, updated_at`
	err := r.DB.QueryRowx(query,
		dto.Email, dto.Username, hashedPassword,
		dto.FirstName, dto.LastName, dto.Role, dto.Telephone,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(id string, dto *UpdateUserDTO) (*User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if dto.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *dto.Email)
		argIdx++
	}
	if dto.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, *dto.Username)
		argIdx++
	}
	if dto.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *dto.FirstName)
		argIdx++
	}
	if dto.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *dto.LastName)
		argIdx++
	}
	if dto.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *dto.Role)
		argIdx++
	}
	if dto.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *dto.IsActive)
		argIdx++
	}
	if dto.Telephone != nil {
		setClauses = append(setClauses, fmt.Sprintf("telephone = $%d", argIdx))
		args = append(args, *dto.Telephone)
		argIdx++
	}

	if len(setClauses) == 0 {
		return r.FindByID(id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d AND deleted_at IS NULL
		 RETURNING id, email, username, password, first_name, last_name, name, role, is_active,
		           refresh_token_hash, telephone, deleted_at, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var user User
	err := r.DB.QueryRowx(query, args...).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) SoftDelete(id string) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
