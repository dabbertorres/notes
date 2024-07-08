package tags

import (
	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/users"
)

type Tag struct {
	ID     uuid.UUID
	Name   string
	Access []users.Access
}

type TagSearchParams struct {
	LastTagID uuid.NullUUID
	Search    string
}
