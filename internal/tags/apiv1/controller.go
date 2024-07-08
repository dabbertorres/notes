package apiv1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/common/apiv1"
	"github.com/dabbertorres/notes/internal/log"
	"github.com/dabbertorres/notes/internal/tags"
	"github.com/dabbertorres/notes/internal/util"
)

type Service interface {
	CreateTag(ctx context.Context, tag *tags.Tag) (*tags.Tag, error)
	UpdateTag(ctx context.Context, tag *tags.Tag) (*tags.Tag, error)
	DeleteTag(ctx context.Context, tagID uuid.UUID) error
	GetTag(ctx context.Context, tagID uuid.UUID) (*tags.Tag, error)
	ListTags(ctx context.Context, params tags.TagSearchParams, pageSize int) (tags []tags.Tag, next *tags.TagSearchParams, err error)
}

func PostTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyDTO, ok := apiv1.ReadJSONOrFail[Tag](w, r)
		if !ok {
			return
		}

		tag, err := bodyDTO.ToDomain()
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewValidationFailureError(err))
			return
		}

		created, err := svc.CreateTag(r.Context(), tag)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		dto := TagFromDomain(created)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &dto)
	}
}

func PutTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tagID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid tag id"))
			return
		}

		body, ok := apiv1.ReadJSONOrFail[WritableTag](w, r)
		if !ok {
			return
		}

		tag, err := body.ToDomain()
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewValidationFailureError(err))
			return
		}

		tag.ID = tagID

		result, err := svc.UpdateTag(r.Context(), tag)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		out := TagFromDomain(result)
		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &out)
	}
}

func DeleteTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tagID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid tag id"))
			return
		}

		if err := svc.DeleteTag(r.Context(), tagID); err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func GetTag(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tagID, err := apiv1.ParsePathValue(r, "id", true, uuid.Parse)
		if err != nil {
			apiv1.WriteError(r.Context(), w, apiv1.NewError(http.StatusBadRequest, "invalid tag id"))
			return
		}

		tag, err := svc.GetTag(r.Context(), tagID)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		dto := TagFromDomain(tag)

		apiv1.WriteJSON(r.Context(), w, http.StatusOK, &dto)
	}
}

func ListTags(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paging, err := parseListTagsParams(r)
		if err != nil {
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		params := tags.TagSearchParams{
			LastTagID: paging.Data.LastTagID,
			Search:    paging.Data.Search,
		}
		results, next, err := svc.ListTags(r.Context(), params, paging.PageSize)
		if err != nil {
			log.Error(r.Context(), "error searching tags", zap.Error(err))
			apiv1.WriteError(r.Context(), w, err)
			return
		}

		page := apiv1.Page[Tag, *ListTagsPageTokenData]{
			NextPageToken: nil,
			Items: util.MapSlice(results, func(tag tags.Tag) Tag {
				return Tag{
					ID:   tag.ID.String(),
					Name: tag.Name,
				}
			}),
		}

		if next != nil {
			page.NextPageToken = &apiv1.PageToken[*ListTagsPageTokenData]{
				Data: &ListTagsPageTokenData{
					LastTagID: next.LastTagID,
					Search:    next.Search,
				},
				PageSize: paging.PageSize,
			}
		}

		apiv1.WriteJSON(r.Context(), w, http.StatusOK, page)
	}
}

func parseListTagsParams(r *http.Request) (token apiv1.PageToken[*ListTagsPageTokenData], err error) {
	rawPageToken := r.FormValue("next_page_token")
	if rawPageToken != "" {
		pageToken, err := apiv1.ParsePageToken[*ListTagsPageTokenData](rawPageToken, 100, 100)
		if err != nil {
			return token, apiv1.NewValidationFailureError(err)
		}

		token = pageToken
	} else {
		token.Data.Search = r.FormValue("search")

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
