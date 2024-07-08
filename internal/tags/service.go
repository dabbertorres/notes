package tags

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/scope"
	"github.com/dabbertorres/notes/internal/users"
)

// TODO: handle error type - e.g. not found, rather than just returning internal server error

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

func (s *Service) CreateTag(ctx context.Context, tag *Tag) (*Tag, error) {
	userID := scope.MustUserID(ctx)

	tagID, err := uuid.NewV7()
	if err != nil {
		return nil, apiv1.StatusError(http.StatusServiceUnavailable)
	}

	tag.ID = tagID
	tag.Access = append(tag.Access, users.Access{
		User:   users.User{ID: userID},
		Access: users.AccessLevelOwner,
	})

	if err := s.repo.SaveTag(ctx, tag); err != nil {
		log.Error(ctx, "error creating tag", zap.Stringer("tag_id", tag.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	return tag, nil
}

func (s *Service) UpdateTag(ctx context.Context, tag *Tag) (*Tag, error) {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersTagAccess(ctx, tag.ID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user tag access", zap.Stringer("tag_id", tag.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	// TODO: if editing access, check if access is owner

	if access < users.AccessLevelEditor {
		return nil, apiv1.StatusError(http.StatusForbidden)
	}

	if err := s.repo.SaveTag(ctx, tag); err != nil {
		log.Error(ctx, "error updating tag", zap.Stringer("tag_id", tag.ID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	return tag, nil
}

func (s *Service) DeleteTag(ctx context.Context, tagID uuid.UUID) error {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersTagAccess(ctx, tagID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user tag access", zap.Stringer("tag_id", tagID), zap.Error(err))
		return apiv1.StatusError(http.StatusInternalServerError)
	}

	// TODO: if editing access, check if access is owner

	if access < users.AccessLevelOwner {
		return apiv1.StatusError(http.StatusForbidden)
	}

	if err := s.repo.DeleteTag(ctx, tagID); err != nil {
		log.Error(ctx, "error deleting tag", zap.Stringer("tag_id", tagID), zap.Error(err))
		return apiv1.StatusError(http.StatusInternalServerError)
	}

	return nil
}

func (s *Service) GetTag(ctx context.Context, tagID uuid.UUID) (*Tag, error) {
	userID := scope.MustUserID(ctx)

	access, err := s.repo.GetUsersTagAccess(ctx, tagID, userID)
	if err != nil {
		log.Error(ctx, "error retrieving user tag access", zap.Stringer("tag_id", tagID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	// TODO: if editing access, check if access is owner

	if access < users.AccessLevelViewer {
		return nil, apiv1.StatusError(http.StatusForbidden)
	}

	tag, err := s.repo.GetTag(ctx, tagID)
	if err != nil {
		log.Error(ctx, "error deleting tag", zap.Stringer("tag_id", tagID), zap.Error(err))
		return nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	return tag, nil
}

func (s *Service) ListTags(ctx context.Context, params TagSearchParams, pageSize int) (tags []Tag, next *TagSearchParams, err error) {
	userID := scope.MustUserID(ctx)

	results, err := s.repo.ListTags(ctx, userID, params, pageSize+1)
	if err != nil {
		log.Error(ctx, "error listing tags", zap.Error(err))
		return nil, nil, apiv1.StatusError(http.StatusInternalServerError)
	}

	if len(results) > pageSize {
		results = results[:pageSize]

		last := &results[len(results)-1]

		next = &TagSearchParams{
			LastTagID: uuid.NullUUID{UUID: last.ID, Valid: true},
			Search:    params.Search,
		}
	}

	return tags, next, err
}
