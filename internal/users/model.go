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
