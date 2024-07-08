package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/database"
	"github.com/dabbertorres/notes/internal/log"
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

func (r *PGXRepository) SaveUser(ctx context.Context, user *User) (out *User, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		params := database.SaveUserParams{
			UserID:     user.ID,
			Name:       user.Name,
			CreatedAt:  pgtype.Timestamptz{Time: user.CreatedAt, Valid: true},
			LastSignIn: pgtype.Timestamptz{Time: user.LastSignIn, Valid: true},
			Active:     user.Active,
		}

		err := r.queries.SaveUser(ctx, tx, params)
		if err != nil {
			log.Error(ctx, "error saving user", zap.Stringer("user_id", user.ID), zap.Error(err))
			return err
		}

		out = user
		return nil
	})

	return out, err
}

func (r *PGXRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	var numDeleted int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) (err error) {
		numDeleted, err = r.queries.DeleteUser(ctx, tx, userID)
		return err
	})
	if err != nil {
		return err
	}

	if numDeleted > 1 {
		panic("uh oh this can't happen")
	}

	if numDeleted == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PGXRepository) GetUser(ctx context.Context, userID uuid.UUID) (out *User, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		user, err := r.queries.GetUser(ctx, tx, userID)
		if err != nil {
			return err
		}

		out = &User{
			ID:         user.UserID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt.Time,
			LastSignIn: user.LastSignIn.Time,
			Active:     user.Active,
		}

		return nil
	})

	return out, err
}
