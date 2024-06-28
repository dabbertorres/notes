package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

type Repository interface {
	SaveUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
}

type Service struct {
	repo Repository
}

func NewService(injector do.Injector) (*Service, error) {
	repo, err := do.InvokeAs[Repository](injector)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo: repo,
	}, nil
}
