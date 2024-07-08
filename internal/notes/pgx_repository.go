package notes

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/do/v2"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/database"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/tags"
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

func (r *PGXRepository) SaveNote(ctx context.Context, note *Note) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		params := database.SaveNoteParams{
			NoteID:    note.ID,
			CreatedAt: pgtype.Timestamptz{Time: note.CreatedAt, Valid: true},
			CreatedBy: uuid.NullUUID{
				UUID:  note.CreatedBy.ID,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{Time: note.UpdatedAt, Valid: true},
			UpdatedBy: uuid.NullUUID{
				UUID:  note.UpdatedBy.ID,
				Valid: true,
			},
			Title: note.Title,
			Body:  note.Body,
		}

		if note.CreatedBy.ID != uuid.Nil {
			params.CreatedBy = uuid.NullUUID{
				UUID:  note.CreatedBy.ID,
				Valid: true,
			}
		}

		if note.UpdatedBy.ID != uuid.Nil {
			params.UpdatedBy = uuid.NullUUID{
				UUID:  note.UpdatedBy.ID,
				Valid: true,
			}
		}

		if err := r.queries.SaveNote(ctx, tx, params); err != nil {
			log.Error(ctx, "error saving note", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		for _, t := range note.Tags {
			params := database.SetNoteTagsParams{
				Column1: uuid.NullUUID{UUID: note.ID, Valid: true},
				Column2: uuid.NullUUID{UUID: t.ID, Valid: true},
			}
			if err := r.queries.SetNoteTags(ctx, tx, params); err != nil {
				log.Error(ctx, "error setting note tags", zap.Stringer("note_id", note.ID), zap.Error(err))
				return apiv1.StatusError(http.StatusInternalServerError)
			}
		}

		for _, a := range note.Access {
			params := database.SetNoteAccessParams{
				Column1: uuid.NullUUID{UUID: a.User.ID, Valid: true},
				Column2: database.NullNotesAccessLevel{
					NotesAccessLevel: database.NotesAccessLevel(a.Access),
					Valid:            a.Access != users.AccessLevelNone,
				},
				Column3: uuid.NullUUID{UUID: note.ID, Valid: true},
			}

			if err := r.queries.SetNoteAccess(ctx, tx, params); err != nil {
				log.Error(ctx, "error setting note access", zap.Stringer("note_id", note.ID), zap.Error(err))
				return apiv1.StatusError(http.StatusInternalServerError)
			}
		}

		return nil
	})
}

func (r *PGXRepository) DeleteNote(ctx context.Context, id uuid.UUID) error {
	var numDeleted int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) (err error) {
		numDeleted, err = r.queries.DeleteNote(ctx, tx, id)
		return err
	})
	if err != nil {
		log.Error(ctx, "error deleting note", zap.Stringer("note_id", id), zap.Error(err))
		return apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	if numDeleted != 1 {
		return apiv1.NewError(http.StatusNotFound, "note does not exist")
	}

	return nil
}

func (r *PGXRepository) GetNote(ctx context.Context, noteID, asUserID uuid.UUID) (note *Note, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		row, err := r.queries.GetNote(ctx, tx, noteID)
		if err != nil {
			log.Error(ctx, "error getting note", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		params := database.GetNoteTagsParams{
			UserID: asUserID,
			NoteID: noteID,
		}

		tagRows, err := r.queries.GetNoteTags(ctx, tx, params)
		if err != nil {
			log.Error(ctx, "error getting note tags", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		accessRows, err := r.queries.GetNoteAccess(ctx, tx, noteID)
		if err != nil {
			log.Error(ctx, "error getting note access", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		var mapAccessErrors []error

		note = &Note{
			ID:        row.NoteID,
			CreatedAt: row.CreatedAt.Time,
			CreatedBy: users.User{ID: row.CreatedBy.UUID},
			UpdatedAt: row.UpdatedAt.Time,
			UpdatedBy: users.User{ID: row.UpdatedBy.UUID},
			Title:     row.Title,
			Body:      row.Body,
			Tags: util.MapSlice(tagRows, func(row database.NotesTag) tags.Tag {
				return tags.Tag{
					ID:   row.TagID,
					Name: row.Name,
				}
			}),
			Access: util.MapSlice(accessRows, func(access database.GetNoteAccessRow) users.Access {
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
			return nil, apiv1.NewError(http.StatusNotFound, "note does not exist")
		}

		log.Error(ctx, "error fetching note", zap.Stringer("note_id", noteID), zap.Error(err))
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return note, nil
}

func (r *PGXRepository) GetUsersNoteAccess(ctx context.Context, noteID, userID uuid.UUID) (level users.AccessLevel, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		accessLevel, err := r.queries.GetUserNoteAccess(ctx, tx, database.GetUserNoteAccessParams{
			NoteID: noteID,
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

func (r *PGXRepository) SearchNotes(ctx context.Context, searchingUser uuid.UUID, search NoteSearchParams, pageSize int) (notes []NoteSearchResult, err error) {
	var searchFunc func(pgx.Tx) error
	switch {
	case search.TagSearch.Valid && search.TextSearch != "":
		params := database.SearchNotesWithTextAndTagParams{
			TextSearch: search.TextSearch,
			UserID:     searchingUser,
			TagID:      search.TagSearch.UUID,
			LastRank:   pgtype.Float4{Float32: search.LastRank, Valid: true},
			PageSize:   int64(pageSize),
		}
		searchFunc = r.searchNotesWithTextAndTag(ctx, params, &notes)

	case search.TagSearch.Valid:
		params := database.SearchNotesWithTagParams{
			UserID:     searchingUser,
			TagID:      search.TagSearch.UUID,
			LastNoteID: search.LastNoteID,
			PageSize:   int64(pageSize),
		}
		searchFunc = r.searchNotesWithTag(ctx, params, &notes)

	case search.TextSearch != "":
		params := database.SearchNotesWithTextParams{
			TextSearch: search.TextSearch,
			UserID:     searchingUser,
			LastRank:   pgtype.Float4{Float32: search.LastRank, Valid: true},
			PageSize:   int64(pageSize),
		}
		searchFunc = r.searchNotesWithText(ctx, params, &notes)

	default:
		params := database.ListNotesParams{
			UserID:     searchingUser,
			LastNoteID: search.LastNoteID,
			PageSize:   int64(pageSize),
		}
		searchFunc = r.listNotes(ctx, params, &notes)
	}

	if err := pgx.BeginFunc(ctx, r.db, searchFunc); err != nil {
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return notes, nil
}

func (r *PGXRepository) searchNotesWithTextAndTag(ctx context.Context, params database.SearchNotesWithTextAndTagParams, notes *[]NoteSearchResult) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		rows, err := r.queries.SearchNotesWithTextAndTag(ctx, tx, params)
		if err != nil {
			return err
		}

		*notes = util.MapSlice(rows,
			func(row database.SearchNotesWithTextAndTagRow) NoteSearchResult {
				return NoteSearchResult{
					ID:      row.NoteID,
					Rank:    row.Rank,
					Title:   row.Title,
					Matched: row.Match.String,
				}
			})

		return nil
	}
}

func (r *PGXRepository) searchNotesWithTag(ctx context.Context, params database.SearchNotesWithTagParams, notes *[]NoteSearchResult) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		rows, err := r.queries.SearchNotesWithTag(ctx, tx, params)
		if err != nil {
			return err
		}

		*notes = util.MapSlice(rows, func(row database.SearchNotesWithTagRow) NoteSearchResult {
			return NoteSearchResult{
				ID:    row.NoteID,
				Title: row.Title,
			}
		})

		return nil
	}
}

func (r *PGXRepository) searchNotesWithText(ctx context.Context, params database.SearchNotesWithTextParams, notes *[]NoteSearchResult) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		rows, err := r.queries.SearchNotesWithText(ctx, tx, params)
		if err != nil {
			return err
		}

		*notes = util.MapSlice(rows, func(row database.SearchNotesWithTextRow) NoteSearchResult {
			return NoteSearchResult{}
		})

		return nil
	}
}

func (r *PGXRepository) listNotes(ctx context.Context, params database.ListNotesParams, notes *[]NoteSearchResult) func(pgx.Tx) error {
	return func(tx pgx.Tx) error {
		rows, err := r.queries.ListNotes(ctx, tx, params)
		if err != nil {
			return err
		}

		*notes = util.MapSlice(rows, func(row database.ListNotesRow) NoteSearchResult {
			return NoteSearchResult{
				ID:    row.NoteID,
				Title: row.Title,
			}
		})

		return nil
	}
}
