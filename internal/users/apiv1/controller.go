package apiv1

import (
	"context"
	"net/http"

	"github.com/dabbertorres/notes/internal/users"
)

type Service interface {
	Create(context.Context, *users.User) (*users.User, error)
	Update(context.Context, *users.User) error
	SignIn(context.Context)
	SignOut(context.Context)
}

type Controller struct {
	svc Service
}

func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
}
