package notes

import (
	"time"

	"github.com/google/uuid"

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
	Tags      []Tag
	Access    []UserAccess
}

type NoteSearchResult struct {
	ID   uuid.UUID
	Rank float32
	// TODO: matched parts
}

type AccessLevel string

const (
	AccessLevelOwner  AccessLevel = "owner"
	AccessLevelEditor AccessLevel = "editor"
	AccessLevelViewer AccessLevel = "viewer"
)

type UserAccess struct {
	User   users.User
	Access AccessLevel
}

type Tag struct {
	ID   uuid.UUID
	User users.User
	Name string
}
