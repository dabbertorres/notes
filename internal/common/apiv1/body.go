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
	if r.Header.Get("Content-Type") != "application/json" {
		WriteError(r.Context(), w, &APIError{
			Status: http.StatusUnsupportedMediaType,
		})
		return
	}

	if accept := r.Header.Get("Accept"); accept != "" && accept != "application/json" {
		WriteError(r.Context(), w, &APIError{
			Status: http.StatusNotAcceptable,
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		WriteError(r.Context(), w, &APIError{
			Status:  http.StatusBadRequest,
			Message: "invalid json",
		})
	} else {
		ok = true
	}

	return value, ok
}

func WriteJSON(ctx context.Context, w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	buf, _ := json.Marshal(body)
	if _, err := w.Write(buf); err != nil {
		log.Info(ctx, "error writing response", zap.Error(err))
	}
}
