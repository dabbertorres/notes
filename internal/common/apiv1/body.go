package apiv1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/log"
)

var ErrFieldNotSet = errors.New("required field not set")

func ParsePathValue[T any](r *http.Request, id string, required bool, parse func(string) (T, error)) (value T, err error) {
	raw := r.PathValue(id)
	if raw == "" {
		if required {
			return value, ErrFieldNotSet
		}

		return value, nil
	}

	value, err = parse(raw)
	return value, err
}

func ParsePathValueWithDefault[T any](r *http.Request, id string, defaultVal T, parse func(string) (T, error)) (value T, err error) {
	raw := r.PathValue(id)
	if raw == "" {
		return defaultVal, nil
	}

	value, err = parse(raw)
	return value, err
}

func ReadJSONOrFail[T any](w http.ResponseWriter, r *http.Request) (value T, ok bool) {
	if contentType := r.Header.Get("Content-Type"); contentType != "" && contentType != "application/json" {
		WriteError(r.Context(), w, &apiError{
			status: http.StatusUnsupportedMediaType,
		})
		return
	}

	if accept := r.Header.Get("Accept"); accept != "" && accept != "application/json" {
		WriteError(r.Context(), w, &apiError{
			status: http.StatusNotAcceptable,
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		WriteError(r.Context(), w, &apiError{
			status: http.StatusBadRequest,
			body: &apiErrorBody{
				Message: "invalid json",
			},
		})
		return value, false
	}

	return value, true
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")

	buf, err := json.Marshal(body)
	if err != nil {
		log.Error(ctx, "failed to marshal response body", zap.Error(err))
		WriteError(ctx, w, &apiError{
			status: http.StatusInternalServerError,
		})
		return
	}

	w.WriteHeader(status)

	if _, err := w.Write(buf); err != nil {
		select {
		case <-ctx.Done():
			log.Debug(ctx, "client closed connection")
		default:
			log.Warn(ctx, "error writing response", zap.Error(err))
		}
	}
}
