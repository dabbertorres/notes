package apiv1

import (
	"time"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/users"
)

type Note struct {
	ID uuid.UUID `json:"id"`
}

func (n *Note) FromDomain(domain *notes.Note) {
	// TODO
}

func (n *Note) ToDomain(id uuid.UUID) *notes.Note {
	return &notes.Note{
		ID: id,
		// TODO
		CreatedAt: time.Time{},
		CreatedBy: &users.User{},
		UpdatedAt: time.Time{},
		UpdatedBy: &users.User{},
		Title:     "",
		Body:      "",
		Tags:      []notes.Tag{},
		Access:    []notes.UserAccess{},
	}
}

type UserAccess struct{}

type Tag struct{}

type User struct{}
