package apiv1

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/notes"
)

type Service interface {
	CreateNote(ctx context.Context, note *notes.Note) (*notes.Note, error)
	UpdateNote(ctx context.Context, note *notes.Note) (*notes.Note, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
	GetNote(ctx context.Context, id uuid.UUID) (*notes.Note, error)
	SearchNotes(ctx context.Context, search string, rank float32) ([]notes.NoteSearchResult, error)
	ListTags(ctx context.Context, nextID, pageSize int) ([]notes.Tag, error)
}

func PostNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		note, ok := apiv1.ReadJSONOrFail[Note](w, r)
		if !ok {
			return
		}

		created, err := svc.CreateNote(r.Context(), note.ToDomain(uuid.Nil))
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		var dto Note
		dto.FromDomain(created)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &dto)
	}
}

func PutNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noteID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, &apiv1.APIError{
				Status:  http.StatusBadRequest,
				Message: "invalid note id",
			})
			return
		}

		note, ok := apiv1.ReadJSONOrFail[Note](w, r)
		if !ok {
			return
		}

		result, err := svc.UpdateNote(r.Context(), note.ToDomain(noteID))
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		var out Note
		out.FromDomain(result)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &out)
	}
}

func DeleteNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetNote(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func ListNotes(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func PostTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func PutTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func DeleteTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func GetTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func ListTags(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}
