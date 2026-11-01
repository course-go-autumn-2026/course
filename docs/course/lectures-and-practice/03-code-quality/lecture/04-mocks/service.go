// Пакет mocksdemo демонстрирует внедрение зависимостей, ручные и сгенерированные моки.
package mocksdemo

import (
	"context"
	"errors"
	"fmt"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=service.go -destination=mocks/user_repository_mock.go -package=mocks

var ErrInvalidUserID = errors.New("invalid user id")

type User struct {
	ID   int
	Name string
}

// UserRepository объявлен на стороне потребителя, которому нужен этот интерфейс.
type UserRepository interface {
	GetUser(ctx context.Context, id int) (User, error)
}

type Service struct {
	repository UserRepository
}

func NewService(repository UserRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetUserName(ctx context.Context, id int) (string, error) {
	if id <= 0 {
		return "", ErrInvalidUserID
	}

	user, err := s.repository.GetUser(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get user %d: %w", id, err)
	}

	return user.Name, nil
}
