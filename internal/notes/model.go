package notes

import (
	"time"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/tags"
	"github.com/dabbertorres/notes/internal/users"
)

type Note struct {
	ID        uuid.UUID
	CreatedAt time.Time
	CreatedBy users.User
	UpdatedAt time.Time
	UpdatedBy users.User
	Title     string
	Body      string
	Tags      []tags.Tag
	Access    []users.Access
}
