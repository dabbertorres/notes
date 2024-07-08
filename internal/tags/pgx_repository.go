package tags

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/database"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

type PGXRepository struct {
	db      database.Database
	queries *database.Queries
}

func NewPGXRepository(injector do.Injector) (*PGXRepository, error) {
	db, err := do.InvokeAs[database.Database](injector)
	if err != nil {
		return nil, err
	}

	return &PGXRepository{
		db:      db,
		queries: database.New(),
	}, nil
}

func (r *PGXRepository) SaveTag(ctx context.Context, tag *Tag) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		params := database.SaveTagParams{
			TagID: tag.ID,
			Name:  tag.Name,
		}
		if err := r.queries.SaveTag(ctx, tx, params); err != nil {
			log.Error(ctx, "error saving tag", zap.Stringer("tag_id", tag.ID), zap.Error(err))
			return err
		}

		for _, a := range tag.Access {
			params := database.SetTagAccessParams{
				Column1: uuid.NullUUID{UUID: a.User.ID, Valid: true},
				Column2: database.NullNotesAccessLevel{
					NotesAccessLevel: database.NotesAccessLevel(a.Access),
					Valid:            a.Access != users.AccessLevelNone,
				},
				Column3: uuid.NullUUID{UUID: tag.ID, Valid: true},
			}

			if err := r.queries.SetTagAccess(ctx, tx, params); err != nil {
				log.Error(ctx, "error setting tag access", zap.Stringer("tag_id", tag.ID), zap.Error(err))
				return err
			}
		}

		return nil
	})
}

func (r *PGXRepository) DeleteTag(ctx context.Context, id uuid.UUID) error {
	var numDeleted int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) (err error) {
		numDeleted, err = r.queries.DeleteTag(ctx, tx, id)
		return err
	})
	if err != nil {
		log.Error(ctx, "error deleting tag", zap.Stringer("tag", id), zap.Error(err))
		return apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	if numDeleted != 1 {
		return apiv1.NewError(http.StatusNotFound, "tag does not exist")
	}

	return nil
}

func (r *PGXRepository) GetTag(ctx context.Context, id uuid.UUID) (tag *Tag, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		row, err := r.queries.GetTag(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "error getting tag", zap.Stringer("tag", id), zap.Error(err))
			return err
		}

		accessRows, err := r.queries.GetTagAccess(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "error getting tag access", zap.Stringer("tag_id", id), zap.Error(err))
			return err
		}

		var mapAccessErrors []error

		tag = &Tag{
			ID:   row.TagID,
			Name: row.Name,
			Access: util.MapSlice(accessRows, func(access database.GetTagAccessRow) users.Access {
				level, err := users.ParseAccessLevel(string(access.Access))
				if err != nil {
					mapAccessErrors = append(mapAccessErrors, err)
				}

				return users.Access{
					User:   users.User{ID: access.UserID},
					Access: level,
				}
			}),
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apiv1.NewError(http.StatusNotFound, "tag does not exist")
		}

		log.Error(ctx, "error fetching tag", zap.Stringer("tag_id", id), zap.Error(err))
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return tag, nil
}

func (r *PGXRepository) ListTags(ctx context.Context, userID uuid.UUID, params TagSearchParams, pageSize int) (tags []Tag, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		rows, err := r.queries.ListTags(ctx, tx, database.ListTagsParams{
			UserID:       userID,
			LastTagID:    params.LastTagID,
			SearchString: params.Search,
			PageSize:     int64(pageSize) + 1,
		})
		if err != nil {
			log.Error(ctx, "error listing tags",
				zap.Stringer("user_id", userID),
				zap.Stringer("last_tag_id", params.LastTagID.UUID),
				zap.Int("fetch_amount", pageSize),
				zap.Error(err),
			)
			return err
		}

		tags = util.MapSlice(rows[:pageSize], func(row database.ListTagsRow) Tag {
			return Tag{
				ID:   row.TagID,
				Name: row.Name,
			}
		})
		return nil
	})
	if err != nil {
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return tags, nil
}

func (r *PGXRepository) GetUsersTagAccess(ctx context.Context, id uuid.UUID, userID uuid.UUID) (level users.AccessLevel, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		accessLevel, err := r.queries.GetUserTagAccess(ctx, tx, database.GetUserTagAccessParams{
			TagID:  id,
			UserID: userID,
		})
		if err != nil {
			return err
		}

		level, err = users.ParseAccessLevel(string(accessLevel))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.AccessLevelNone, nil
		}

		return level, err
	}

	return level, nil
}
