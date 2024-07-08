package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Name       string
	CreatedAt  time.Time
	LastSignIn time.Time
	Active     bool
}

//go:generate go run ../../tools/stringer -type=AccessLevel -trimprefix=AccessLevel -linecomment -lower
type AccessLevel byte

const (
	AccessLevelNone AccessLevel = iota
	AccessLevelViewer
	AccessLevelEditor
	AccessLevelOwner
)

type Access struct {
	User   User
	Access AccessLevel
}
