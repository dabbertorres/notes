package apiv1

import (
	"context"
	"net/http"

	"github.com/dabbertorres/notes/internal/users"
)

type TODO = any

type Service interface {
	Create(context.Context, *users.User) (*users.User, error)
	Update(context.Context, *users.User) error
	SignIn(context.Context, TODO) error
	SignOut(context.Context, TODO) error
}

func PostUser(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func PutUser(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func PostSession(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteSession(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}
