package tags

import (
	"context"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/users"
)

type Repository interface {
	SaveTag(ctx context.Context, tag *Tag) error
	DeleteTag(ctx context.Context, id uuid.UUID) error
	GetTag(ctx context.Context, id uuid.UUID) (*Tag, error)
	ListTags(ctx context.Context, userID uuid.UUID, params TagSearchParams, pageSize int) ([]Tag, error)
	GetUsersTagAccess(ctx context.Context, id uuid.UUID, userID uuid.UUID) (users.AccessLevel, error)
}
