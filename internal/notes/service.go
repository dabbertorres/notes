package notes

import (
	"cmp"
	"context"
	"math"
	"net/http"
	"slices"
	"time"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/database"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/scope"
	"github.com/dabbertorres/notes/util"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type Repository interface {
	SaveNote(ctx context.Context, note *Note, removeAccess []UserAccess, removeTags []Tag) (*Note, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
	GetNote(ctx context.Context, id uuid.UUID) (*Note, error)
	SearchNotes(ctx context.Context, searchingUser uuid.UUID, search string, rank float32) ([]NoteSearchResult, error)
	ListTags(ctx context.Context, userID uuid.UUID, nextID, fetchAmount int) ([]Tag, error)
}

type Service struct {
	repo Repository
}

func NewService(injector *do.Injector) (*Service, error) {
	db := do.MustInvoke[database.Database](injector)
	return &Service{
		repo: newRepository(db),
	}, nil
}

func (s *Service) CreateNote(ctx context.Context, note *Note) (*Note, error) {
	user, ok := scope.User(ctx)
	if !ok {
		log.Info(ctx, "user missing from request context")
		return nil, &apiv1.APIError{
			Status: http.StatusForbidden,
		}
	}

	note.ID = uuid.New()
	note.CreatedAt = time.Now()
	note.CreatedBy = &user
	note.UpdatedAt = note.CreatedAt
	note.UpdatedBy = &user

	note, err := s.repo.SaveNote(ctx, note, nil, nil)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) UpdateNote(ctx context.Context, note *Note) (*Note, error) {
	currentState, err := s.repo.GetNote(ctx, note.ID)
	if err != nil {
		return nil, err
	}

	sortAccess := func(a, b UserAccess) int { return cmp.Compare(a.User.ID.String(), b.User.ID.String()) }
	sortTags := func(a, b Tag) int { return cmp.Compare(a.ID.String(), b.ID.String()) }

	slices.SortFunc(currentState.Access, sortAccess)
	slices.SortFunc(currentState.Tags, sortTags)

	slices.SortFunc(note.Access, sortAccess)
	slices.SortFunc(note.Tags, sortTags)

	_, removedAccess := util.SliceDiffBy(currentState.Access, note.Access, func(lhs, rhs UserAccess) bool {
		return lhs.User.ID == rhs.User.ID
	})

	_, removedTags := util.SliceDiffBy(currentState.Tags, note.Tags, func(lhs, rhs Tag) bool {
		return lhs.ID == rhs.ID
	})

	return s.repo.SaveNote(ctx, note, removedAccess, removedTags)
}

func (s *Service) DeleteNote(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteNote(ctx, id)
}

func (s *Service) GetNote(ctx context.Context, id uuid.UUID) (*Note, error) {
	return s.repo.GetNote(ctx, id)
}

func (s *Service) SearchNotes(ctx context.Context, search string, rank float32) ([]NoteSearchResult, error) {
	user, ok := scope.User(ctx)
	if !ok {
		log.Info(ctx, "user missing from request context")
		return nil, &apiv1.APIError{
			Status: http.StatusForbidden,
		}
	}

	if rank == 0 {
		rank = math.MaxFloat32
	}

	return s.repo.SearchNotes(ctx, user.ID, search, rank)
}

func (s *Service) ListTags(ctx context.Context, nextID, pageSize int) ([]Tag, error) {
	user, ok := scope.User(ctx)
	if !ok {
		log.Info(ctx, "user missing from request context")
		return nil, &apiv1.APIError{
			Status: http.StatusForbidden,
		}
	}

	return s.repo.ListTags(ctx, user.ID, nextID, pageSize)
}
