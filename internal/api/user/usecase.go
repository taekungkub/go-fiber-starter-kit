package user

import (
	"errors"
	"go-fiber-stater-kit/pkg/common"
	"go-fiber-stater-kit/pkg/core"
)

type UseCase interface {
	FindAll(page, limit int64, sort, order string) (*core.Paging, error)
	FindByID(id string) (*UserResponse, error)
	Create(dto *CreateUserDTO) (*UserResponse, error)
	Update(id string, dto *UpdateUserDTO) (*UserResponse, error)
	Delete(id string) error
}

type useCase struct {
	Repo Repository
	// RedisClient *redis.Client
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{
		Repo: repo,
		// RedisClient: redisClient,
	}
}

func (u *useCase) FindAll(page, limit int64, sort, order string) (*core.Paging, error) {
	var offset int64
	if limit > 0 {
		offset = core.Offset(page, limit)
	}

	users, err := u.Repo.FindAll(limit, offset, sort, order)
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}

	count, err := u.Repo.Count()
	if err != nil {
		return nil, errors.New("failed to count users")
	}

	// Convert to response
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	if limit <= 0 {
		// No pagination — return all with count
		paging := core.Paging{
			List:  responses,
			Page:  1,
			Limit: count,
			Count: 1,
			Total: count,
			Start: 0,
			End:   count - 1,
		}
		return &paging, nil
	}

	paging := core.Pagination(page, limit,
		func() int64 { return count },
		func(limit int64, offset int64) interface{} { return responses },
	)

	return &paging, nil
}

func (u *useCase) FindByID(id string) (*UserResponse, error) {
	// Try cache first
	// ctx := context.Background()
	// cached, err := u.RedisClient.Get(ctx, "user:detail:"+id).Result()
	// if err == nil {
	// 	var resp UserResponse
	// 	if json.Unmarshal([]byte(cached), &resp) == nil {
	// 		return &resp, nil
	// 	}
	// }

	user, err := u.Repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	resp := user.ToResponse()

	// Cache the result
	// if data, err := json.Marshal(resp); err == nil {
	// 	u.RedisClient.Set(ctx, "user:detail:"+id, data, 10*time.Minute)
	// }

	return &resp, nil
}

func (u *useCase) Create(dto *CreateUserDTO) (*UserResponse, error) {
	hashedPassword, err := common.HashPassword(dto.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user, err := u.Repo.Create(dto, hashedPassword)
	if err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (u *useCase) Update(id string, dto *UpdateUserDTO) (*UserResponse, error) {
	user, err := u.Repo.Update(id, dto)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	// Invalidate cache
	// ctx := context.Background()
	// u.RedisClient.Del(ctx, "user:detail:"+id)

	resp := user.ToResponse()
	return &resp, nil
}

func (u *useCase) Delete(id string) error {
	if err := u.Repo.SoftDelete(id); err != nil {
		return errors.New("failed to delete user")
	}

	// Invalidate cache
	// ctx := context.Background()
	// u.RedisClient.Del(ctx, "user:detail:"+id)

	return nil
}
