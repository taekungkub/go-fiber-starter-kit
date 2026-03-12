package ask

import (
	"errors"
)

type UseCase interface {
	Create(dto *CreateChatDTO) (string, error)
}

type useCase struct {
	Repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{
		Repo: repo,
	}
}

func (u *useCase) Create(dto *CreateChatDTO) (string, error) {

	answer, err := u.Repo.Create(dto)
	if err != nil {
		return "", errors.New("failed to create user: " + err.Error())
	}

	return answer, nil
}
