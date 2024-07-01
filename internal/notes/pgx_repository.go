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
	notesdb "github.com/dabbertorres/notes/internal/notes/db"
	"github.com/dabbertorres/notes/internal/users"
	"github.com/dabbertorres/notes/internal/util"
)

// TODO: access control

type PGXRepository struct {
	db      database.Database
	queries *notesdb.Queries
}

func NewPGXRepository(injector do.Injector) (*PGXRepository, error) {
	db, err := do.InvokeAs[database.Database](injector)
	if err != nil {
		return nil, err
	}

	return &PGXRepository{
		db:      db,
		queries: notesdb.New(),
	}, nil
}

func (r *PGXRepository) SaveNote(ctx context.Context, note *Note, removeAccess []UserAccess, removeTags []Tag) (out *Note, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		params := notesdb.SaveNoteParams{
			NoteID:    note.ID,
			CreatedAt: pgtype.Timestamptz{Time: note.CreatedAt, Valid: true},
			CreatedBy: uuid.NullUUID{}, // TODO
			UpdatedAt: pgtype.Timestamptz{Time: note.UpdatedAt, Valid: true},
			UpdatedBy: uuid.NullUUID{}, // TODO
			Title:     note.Title,
			Body:      note.Body,
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

		result, err := r.queries.SaveNote(ctx, tx, params)
		if err != nil {
			log.Error(ctx, "error saving note", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		out = &Note{
			ID:        result.NoteID,
			CreatedAt: result.CreatedAt.Time,
			CreatedBy: users.User{
				ID: result.CreatedBy.UUID,
			},
			UpdatedAt: result.UpdatedAt.Time,
			UpdatedBy: users.User{
				ID: result.UpdatedBy.UUID,
			},
			Title:  result.Title,
			Body:   result.Title,
			Tags:   nil,
			Access: nil,
		}

		err = r.queries.AddNoteTags(ctx, tx,
			util.MapSlice(note.Tags, func(t Tag) notesdb.AddNoteTagsParams {
				return notesdb.AddNoteTagsParams{
					NoteID: note.ID,
					TagID:  t.ID,
				}
			})).Close()
		if err != nil {
			log.Error(ctx, "error adding note tags", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		err = r.queries.DeleteNoteTags(ctx, tx,
			util.MapSlice(removeTags, func(t Tag) notesdb.DeleteNoteTagsParams {
				return notesdb.DeleteNoteTagsParams{
					NoteID: note.ID,
					TagID:  t.ID,
				}
			})).Close()
		if err != nil {
			log.Error(ctx, "error deleting note tags", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		err = r.queries.AddNoteAccess(ctx, tx,
			util.MapSlice(note.Access, func(access UserAccess) notesdb.AddNoteAccessParams {
				return notesdb.AddNoteAccessParams{
					NoteID: note.ID,
					UserID: access.User.ID,
					Access: notesdb.NotesAccessLevel(access.Access),
				}
			})).Close()
		if err != nil {
			log.Error(ctx, "error adding note access", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		err = r.queries.DeleteNoteAccess(ctx, tx,
			util.MapSlice(removeAccess, func(access UserAccess) notesdb.DeleteNoteAccessParams {
				return notesdb.DeleteNoteAccessParams{
					NoteID: note.ID,
					UserID: access.User.ID,
				}
			})).Close()
		if err != nil {
			log.Error(ctx, "error deleting note access", zap.Stringer("note_id", note.ID), zap.Error(err))
			return err
		}

		return nil
	})
	if err != nil {
		log.Error(ctx, "error saving note", zap.Stringer("note_id", note.ID), zap.Error(err))
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return out, nil
}

func (r *PGXRepository) DeleteNote(ctx context.Context, id uuid.UUID) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		return r.queries.DeleteNote(ctx, tx, id)
	})
	if err != nil {
		log.Error(ctx, "error deleting note", zap.Stringer("note_id", id), zap.Error(err))
		return apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return nil
}

func (r *PGXRepository) GetNote(ctx context.Context, noteID uuid.UUID) (note *Note, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		row, err := r.queries.GetNote(ctx, tx, noteID)
		if err != nil {
			log.Error(ctx, "error getting note", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		note = &Note{
			ID:        row.NoteID,
			CreatedAt: row.CreatedAt.Time,
			CreatedBy: users.User{},
			UpdatedAt: row.UpdatedAt.Time,
			UpdatedBy: users.User{},
			Title:     row.Title,
			Body:      row.Body,
		}

		if row.CreatedBy.Valid {
			note.CreatedBy = users.User{
				ID: row.CreatedBy.UUID,
			}
		}

		if row.UpdatedBy.Valid {
			note.UpdatedBy = users.User{
				ID: row.UpdatedBy.UUID,
			}
		}

		tagRows, err := r.queries.GetNoteTags(ctx, tx, noteID)
		if err != nil {
			log.Error(ctx, "error getting note tags", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		note.Tags = util.MapSlice(tagRows, func(row notesdb.GetNoteTagsRow) Tag {
			return Tag{
				ID:   row.TagID,
				User: users.User{ID: row.UserID},
				Name: row.Name,
			}
		})

		accessRows, err := r.queries.GetNoteAccess(ctx, tx, noteID)
		if err != nil {
			log.Error(ctx, "error getting note access", zap.Stringer("note_id", noteID), zap.Error(err))
			return err
		}

		note.Access = util.MapSlice(accessRows, func(access notesdb.GetNoteAccessRow) UserAccess {
			return UserAccess{
				User:   users.User{ID: access.UserID},
				Access: AccessLevel(access.Access),
			}
		})

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

func (r *PGXRepository) SearchNotes(ctx context.Context, searchingUser uuid.UUID, search string, pageSize int) (notes []NoteSearchResult, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		rows, err := r.queries.SearchNotes(ctx, tx, notesdb.SearchNotesParams{
			Search:   search,
			PageSize: int64(pageSize),
		})
		if err != nil {
			log.Error(ctx, "error searching notes",
				zap.Error(err),
			)
			return err
		}

		notes = util.MapSlice(rows, func(row notesdb.SearchNotesRow) NoteSearchResult {
			return NoteSearchResult{
				ID:   row.NoteID,
				Rank: row.Rank,
			}
		})

		return nil
	})
	if err != nil {
		return nil, apiv1.NewError(http.StatusInternalServerError, "try again later")
	}

	return notes, nil
}

func (r *PGXRepository) ListTags(ctx context.Context, userID uuid.UUID, nextID, pageSize int) (tags []Tag, err error) {
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		rows, err := r.queries.ListTags(ctx, tx, notesdb.ListTagsParams{
			UserID:   userID,
			PageSize: int64(pageSize),
		})
		if err != nil {
			log.Error(ctx, "error listing tags",
				zap.Stringer("user_id", userID),
				zap.Int("next_id", nextID),
				zap.Int("fetch_amount", pageSize),
				zap.Error(err),
			)
			return err
		}

		tags = util.MapSlice(rows, func(row notesdb.ListTagsRow) Tag {
			return Tag{
				ID:   row.TagID,
				User: users.User{ID: userID},
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
