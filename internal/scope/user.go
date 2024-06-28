package scope

import (
	"context"

	"github.com/google/uuid"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	user, ok := ctx.Value(userIDKey{}).(uuid.UUID)
	return user, ok
}
