package apiv1

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

type User struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	LastSignIn string `json:"last_sign_in"`
}

func (u *User) ToDomain() (*users.User, error) {
	var errs []error

	out := &users.User{
		ID:         apiv1.Validate(".id", u.ID, &errs, uuid.Parse),
		Name:       u.Name,
		CreatedAt:  apiv1.Validate(".created_at", u.CreatedAt, &errs, apiv1.ParseRFC3339),
		LastSignIn: apiv1.Validate(".last_sign_in", u.LastSignIn, &errs, apiv1.ParseRFC3339),
	}

	if len(errs) != 0 {
		return nil, apiv1.NewError(
			http.StatusBadRequest,
			"one or more invalid fields",
			util.MapSlice(errs, error.Error)...,
		)
	}

	return out, nil
}

func (u *User) FromDomain(in *users.User) {
	u.ID = in.ID.String()
	u.Name = in.Name
	u.CreatedAt = in.CreatedAt.Format(time.RFC3339)
	u.LastSignIn = in.LastSignIn.Format(time.RFC3339)
}
