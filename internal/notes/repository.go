package notes

import (
	"context"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/users"
)

type Repository interface {
	SaveNote(ctx context.Context, note *Note) error
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
	GetNote(ctx context.Context, noteID, asUserID uuid.UUID) (*Note, error)
	GetUsersNoteAccess(ctx context.Context, noteID, userID uuid.UUID) (users.AccessLevel, error)
	SearchNotes(ctx context.Context, asUserID uuid.UUID, search NoteSearchParams, pageSize int) ([]NoteSearchResult, error)
}

type NoteSearchParams struct {
	TextSearch string
	TagSearch  uuid.NullUUID
	LastNoteID uuid.NullUUID
	LastRank   float32
}

type NoteSearchResult struct {
	ID      uuid.UUID
	Rank    float32
	Title   string
	Matched string
}
