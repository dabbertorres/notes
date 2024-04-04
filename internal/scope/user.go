package scope

import (
	"context"

	"github.com/dabbertorres/notes/internal/users"
)

type userKey struct{}

func WithUser(ctx context.Context, user users.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func User(ctx context.Context) (users.User, bool) {
	user, ok := ctx.Value(userKey{}).(users.User)
	return user, ok
}
