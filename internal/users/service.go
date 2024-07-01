package users

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/samber/do/v2"

	"github.com/dabbertorres/notes/internal/common/apiv1"
)

type TODO = any

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

func (s *Service) Create(context.Context, *User) (*User, error) {
	return nil, apiv1.StatusError(http.StatusNotImplemented)
}

func (s *Service) Update(context.Context, *User) error {
	return apiv1.StatusError(http.StatusNotImplemented)
}

func (s *Service) SignIn(context.Context, TODO) error {
	return apiv1.StatusError(http.StatusNotImplemented)
}

func (s *Service) SignOut(context.Context, TODO) error {
	return apiv1.StatusError(http.StatusNotImplemented)
}
