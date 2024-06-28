package apiv1

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/dabbertorres/notes/internal/log"
)

// Error is the minimal interface for error types that should be used to respond to a request, using [WriteError].
//
// No response body will be written if this is the only interface implemented by an error, meaning the Error() string
// method will _not_ be called.
// This is done to minimize the possibility of leaking internal errors in responses.
//
// This also allows for error types to report both internal (via Error() string) and external (via Body() any) details.
//
// To include a response body (e.g. for reporting error reasons), see [ErrorWithBody].
type Error interface {
	Status() int
	error
}

// ErrorWithBody is the interface that should be implemented by error types that wish to include a response body.
// Note that [Body] can still return nil, in which case no response body will be written.
//
// Also note that [Body] must return a value that can be marshaled to JSON.
type ErrorWithBody interface {
	Error
	Body() any
}

// NewError creates a generic new [Error] from the given information.
func NewError(status int, message string, details ...string) error {
	return &apiError{
		status: status,
		body: &apiErrorBody{
			Message: message,
			Details: details,
		},
	}
}

func getErrorBody(err Error) any {
	if wb, ok := err.(ErrorWithBody); ok {
		return wb.Body()
	}
	return nil
}

// StatusError is an error that communicates an HTTP status and nothing more.
type StatusError int

func (e StatusError) Status() int   { return int(e) }
func (e StatusError) Error() string { return http.StatusText(int(e)) }

// apiError contains the data to write an error response to the wire.
type apiError struct {
	body   *apiErrorBody
	status int
}

type apiErrorBody struct {
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *apiError) Status() int { return e.status }
func (e *apiError) Body() any   { return e.body }
func (e *apiError) Error() string {
	if e.body != nil && e.body.Message != "" {
		return e.body.Message
	}
	return http.StatusText(e.status)
}

func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	var (
		apiErr         Error
		apiErrWithBody ErrorWithBody
	)
	switch {
	case errors.As(err, &apiErrWithBody):
	case errors.As(err, &apiErr):
	default:
		log.Warn(ctx, "non-apiv1.Error leaked to handler", zap.Error(err))
		apiErr = &apiError{
			status: http.StatusInternalServerError,
			body: &apiErrorBody{
				Message: "internal error, try again later",
			},
		}
	}

	if body := getErrorBody(apiErr); body != nil {
		WriteJSON(ctx, w, apiErr.Status(), apiErr)
	} else {
		w.WriteHeader(apiErr.Status())
	}
}
