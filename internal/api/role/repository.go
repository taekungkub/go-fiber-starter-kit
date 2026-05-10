package role

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindAll(limit, offset int64, sort, order string) ([]Role, error)
	Count() (int64, error)
	FindByID(id string) (*Role, error)
	Create(dto *CreateRoleDTO) error
	UpdateRole(dto *UpdateRoleDTO) error
	SoftDelete(id string) error
}

type repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{DB: db}
}

func (r *repository) FindAll(limit, offset int64, sort, order string) ([]Role, error) {
	var roles []Role

	query := fmt.Sprintf(
		`SELECT id, role_name, description, status, created_at, updated_at
		 FROM roles WHERE deleted_at IS NULL
		 ORDER BY %s %s`, sort, order,
	)

	var err error
	if limit > 0 {
		query += " LIMIT $1 OFFSET $2"
		err = r.DB.Select(&roles, query, limit, offset)
	} else {
		// No limit — return all records
		err = r.DB.Select(&roles, query)
	}

	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []Role{}
	}

	return roles, nil
}

func (r *repository) Count() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`
	err := r.DB.Get(&count, query)
	return count, err
}

func (r *repository) FindByID(id string) (*Role, error) {
	var role Role
	query := `SELECT id, role_name, description, status, created_at, updated_at
	          FROM roles WHERE id = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&role, query, id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) Create(dto *CreateRoleDTO) error {
	var role Role
	query := `
		INSERT INTO tbm_roles (role_name, description, status)
		VALUES (:role_name, :description, :status)
		RETURNING id
	`

	rows, err := r.DB.NamedQuery(query, role)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&role.Id)
	}
	return nil
}

func (r *repository) UpdateRole(dto *UpdateRoleDTO) error {
	query := `
		UPDATE tbm_roles
		SET role_name      = :role_name,
		    description    = :description,
		    "departmentId" = :departmentId,
		    status         = :status,
		    "updatedBy"    = :updatedBy,
		    "updatedAt"    = NOW()
		WHERE id = :id
	`
	_, err := r.DB.NamedExec(query, dto)
	return err
}

func (r *repository) SoftDelete(id string) error {
	query := `UPDATE roles SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}
