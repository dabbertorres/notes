package apiv1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/notes"
	"github.com/dabbertorres/notes/internal/util"
)

type Service interface {
	CreateNote(ctx context.Context, note *notes.Note) (*notes.Note, error)
	UpdateNote(ctx context.Context, note *notes.Note) (*notes.Note, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
	GetNote(ctx context.Context, id uuid.UUID) (*notes.Note, error)
	SearchNotes(ctx context.Context, params notes.NoteSearchParams, pageSize int) (results []notes.NoteSearchResult, next *notes.NoteSearchParams, err error)
}

func PostNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyDto, ok := apiv1.ReadJSONOrFail[WritableNote](w, r)
		if !ok {
			return
		}

		note, err := bodyDto.ToDomain()
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewValidationFailureError(err))
			return
		}

		created, err := svc.CreateNote(r.Context(), note)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		dto := NoteFromDomain(created)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &dto)
	}
}

func PutNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noteID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid note id"))
			return
		}

		body, ok := apiv1.ReadJSONOrFail[WritableNote](w, r)
		if !ok {
			return
		}

		note, err := body.ToDomain()
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewValidationFailureError(err))
			return
		}

		note.ID = noteID

		result, err := svc.UpdateNote(r.Context(), note)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		out := NoteFromDomain(result)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &out)
	}
}

func DeleteNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noteID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid note id"))
			return
		}

		if err := svc.DeleteNote(r.Context(), noteID); err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func GetNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noteID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid note id"))
			return
		}

		note, err := svc.GetNote(r.Context(), noteID)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		dto := NoteFromDomain(note)

		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &dto)
	}
}

func ListNotes(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paging, err := parseListNotesParams(r)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		params := notes.NoteSearchParams{
			TextSearch: paging.Data.TextSearch,
			TagSearch:  paging.Data.TagSearch,
			LastNoteID: paging.Data.LastNoteID,
			LastRank:   paging.Data.LastRank,
		}
		results, next, err := svc.SearchNotes(r.Context(), params, paging.PageSize)
		if err != nil {
			log.Error(r.Context(), "error searching notes", zap.Error(err))
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		page := apiv1.Page[Note, *ListNotesPageTokenData]{
			NextPageToken: nil,
			Items: util.MapSlice(results, func(note notes.NoteSearchResult) Note {
				return Note{ID: note.ID.String()}
			}),
		}

		if next != nil {
			page.NextPageToken = &apiv1.PageToken[*ListNotesPageTokenData]{
				Data: &ListNotesPageTokenData{
					LastNoteID: next.LastNoteID,
					LastRank:   next.LastRank,
					TextSearch: next.TextSearch,
					TagSearch:  next.TagSearch,
				},
				PageSize: paging.PageSize,
			}
		}

		apiv1.WriteJSON(r.Context(), w, http.StatusOK, page)
	}
}

func parseListNotesParams(r *http.Request) (token apiv1.PageToken[*ListNotesPageTokenData], err error) {
	rawPageToken := r.FormValue("next_page_token")
	if rawPageToken != "" {
		pageToken, err := apiv1.ParsePageToken[*ListNotesPageTokenData](rawPageToken, 100, 100)
		if err != nil {
			return token, apiv1.NewValidationFailureError(err)
		}

		token = pageToken
	} else {
		token.Data.TextSearch = r.FormValue("text")

		if tag := r.FormValue("tag"); tag != "" {
			tagID, err := uuid.Parse(tag)
			if err != nil {
				return token, apiv1.NewValidationFailureError(err)
			}

			token.Data.TagSearch.UUID = tagID
			token.Data.TagSearch.Valid = true
		}

		if size := r.FormValue("page_size"); size != "" {
			pageSize, err := strconv.ParseInt(size, 10, 64)
			if err != nil {
				return token, apiv1.NewValidationFailureError(err)
			}

			token.PageSize = int(pageSize)
		}
	}

	return token, nil
}
