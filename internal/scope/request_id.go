package scope

import (
	"context"

	"github.com/google/uuid"
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(requestIDKey{}).(uuid.UUID)
	return v
}
