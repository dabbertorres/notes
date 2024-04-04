package apiv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dabbertorres/notes/internal/log"
	"go.uber.org/zap"
)

type APIError struct {
	Status  int      `json:"-"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

type InvalidFieldError struct {
	Field string `json:"field"`
	Err   error  `json:"error"`
}

func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		log.Warn(ctx, "non-APIError leaked to handler", zap.Error(err))
		apiErr = &APIError{
			Status:  http.StatusInternalServerError,
			Message: "internal error, try again later",
		}
	}

	w.WriteHeader(apiErr.Status)

	if len(apiErr.Message) != 0 {
		body, _ := json.Marshal(apiErr)
		w.Write(body)
	}
}
