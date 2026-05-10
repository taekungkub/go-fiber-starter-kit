package role

import (
	"errors"
	"go-fiber-stater-kit/pkg/core"
)

type UseCase interface {
	FindAll(page, limit int64, sort, order string) (*core.Paging, error)
	FindByID(id string) (*RoleResponse, error)
	Create(dto *CreateRoleDTO) (*RoleResponse, error)
	Update(id string, dto *UpdateRoleDTO) (*RoleResponse, error)
	Delete(id string) error
}

type useCase struct {
	Repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{
		Repo: repo,
	}
}

func (u *useCase) FindAll(page, limit int64, sort, order string) (*core.Paging, error) {
	var offset int64
	if limit > 0 {
		offset = core.Offset(page, limit)
	}

	roles, err := u.Repo.FindAll(limit, offset, sort, order)
	if err != nil {
		return nil, errors.New("failed to fetch roles")
	}

	count, err := u.Repo.Count()
	if err != nil {
		return nil, errors.New("failed to count roles")
	}

	// Convert to response
	responses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = role.ToResponse()
	}

	if limit <= 0 {
		// No pagination — return all with count
		paging := core.Paging{
			List:      responses,
			Page:      1,
			Limit:     count,
			TotalPage: 1,
			Total:     count,
		}
		return &paging, nil
	}

	paging := core.Pagination(page, limit,
		func() int64 { return count },
		func(limit int64, offset int64) interface{} { return responses },
	)

	return &paging, nil
}

func (u *useCase) FindByID(id string) (*RoleResponse, error) {

	role, err := u.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	resp := role.ToResponse()

	return &resp, nil
}

func (u *useCase) Create(dto *CreateRoleDTO) (*RoleResponse, error) {
	err := u.Repo.Create(dto)
	if err != nil {
		return nil, errors.New("failed to create role: " + err.Error())
	}

	resp := RoleResponse{}
	return &resp, nil
}

func (u *useCase) Update(id string, dto *UpdateRoleDTO) (*RoleResponse, error) {
	err := u.Repo.UpdateRole(dto)
	if err != nil {
		return nil, errors.New("failed to update role")
	}

	resp := RoleResponse{}
	return &resp, nil
}

func (u *useCase) Delete(id string) error {
	if err := u.Repo.SoftDelete(id); err != nil {
		return errors.New("failed to delete role")
	}

	return nil
}
