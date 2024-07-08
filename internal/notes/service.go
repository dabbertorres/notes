package notes

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/scope"
	"github.com/dabbertorres/notes/internal/users"
)

type Service struct {
	repo Repository
}

func NewService(injector do.Injector) (*Service, error) {
	repo, err := do.InvokeAs[Repository](injector)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo: repo,
	}, nil
}

func (s *Service) CreateNote(ctx context.Context, note *Note) (*Note, error) {
	userID := scope.MustUserID(ctx)

	noteID, err := uuid.NewV7()
	if err != nil {
		return nil, apiv1.StatusError(http.StatusServiceUnavailable)
	}

	note.ID = noteID
	note.CreatedAt = time.Now()
	note.CreatedBy = users.User{ID: userID}
	note.UpdatedAt = note.CreatedAt
	note.UpdatedBy = users.User{ID: userID}
	note.Access = append(note.Access, users.Access{
		User:   note.CreatedBy,
		Access: users.AccessLevelOwner,
	})

	if err := s.repo.SaveNote(ctx, note); err != nil {
		log.Error(ctx, "error creating note", zap.Stringer("note_id", note.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	return note, nil
}

func (s *Service) UpdateNote(ctx context.Context, note *Note) (*Note, error) {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersNoteAccess(ctx, note.ID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user note access", zap.Stringer("note_id", note.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	// TODO: if editing access, check if access is owner

	if access < users.AccessLevelEditor {
		return nil, apiv1.StatusError(http.StatusForbidden)
	}

	note.UpdatedAt = time.Now()
	note.UpdatedBy.ID = userID

	if err := s.repo.SaveNote(ctx, note); err != nil {
		log.Error(ctx, "error saving note", zap.Stringer("note_id", note.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	return note, nil
}

func (s *Service) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user note access", zap.Stringer("note_id", noteID), zap.Error(err))
		return apiv1.StatusError(http.StatusInternalServerError)
	}

	if access < users.AccessLevelOwner {
		return apiv1.StatusError(http.StatusForbidden)
	}

	return s.repo.DeleteNote(ctx, noteID)
}

func (s *Service) GetNote(ctx context.Context, noteID uuid.UUID) (*Note, error) {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user note access", zap.Stringer("note_id", noteID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	if access < users.AccessLevelViewer {
		return nil, apiv1.StatusError(http.StatusForbidden)
	}

	return s.repo.GetNote(ctx, noteID, userID)
}

func (s *Service) SearchNotes(ctx context.Context, params NoteSearchParams, pageSize int) ([]NoteSearchResult, *NoteSearchParams, error) {
	userID := scope.MustUserID(ctx)

	if pageSize == 0 {
		pageSize = 100
	}

	// retrive one more to see if there is another page to fetch
	results, err := s.repo.SearchNotes(ctx, userID, params, pageSize+1)
	if err != nil {
		log.Error(ctx, "error searching notes", zap.Error(err))
		return nil, nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	var next *NoteSearchParams
	if len(results) > pageSize {
		results = results[:pageSize]

		last := &results[len(results)-1]

		next = &NoteSearchParams{
			TextSearch: params.TextSearch,
			TagSearch:  params.TagSearch,
			LastNoteID: uuid.NullUUID{UUID: last.ID, Valid: true},
			LastRank:   last.Rank,
		}
	}

	return results, next, err
}
