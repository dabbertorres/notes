package apiv1

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dabbertorres/notes/internal/util"
)

// ValidationFailureError is used for sending as a response to an API call that failed validation.
//
// It should be created with [NewValidationFailureError].
type ValidationFailureError struct {
	body validationFailureErrorBody
}

type validationFailureErrorBody struct {
	Message string  `json:"message"`
	Details []error `json:"details"`
}

func NewValidationFailureError(err error) *ValidationFailureError {
	var details []error
	if joinedErrs, ok := err.(interface{ Unwrap() []error }); ok {
		details = joinedErrs.Unwrap()
	} else {
		details = []error{err}
	}

	return &ValidationFailureError{
		body: validationFailureErrorBody{
			Message: "one or more invalid fields in request",
			Details: details,
		},
	}
}

func (e *ValidationFailureError) Status() int   { return http.StatusBadRequest }
func (e *ValidationFailureError) Body() any     { return e.body }
func (e *ValidationFailureError) Error() string { return e.body.Message }

// InvalidFieldError should be returned whenever a field fails validation/parsing.
//
// It is typically used in a slice for the error details of [ValidationFailureError].
type InvalidFieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

func (e *InvalidFieldError) Status() int   { return http.StatusBadRequest }
func (e *InvalidFieldError) Body() any     { return e }
func (e *InvalidFieldError) Error() string { return fmt.Sprintf("%s: %v", e.Field, e.Err) }

func (e *InvalidFieldError) Qualify(parent string) *InvalidFieldError {
	e.Field = parent + e.Field
	return e
}

// Validate calls parse with in, and if successful, returns the result.
// If parse returns an error, the error is appended to errs, and the zero value of O is returned.
//
// This signature makes it easy to map between types in a nice manner, while recording errors.
// For example:
//
//	 type Foo struct {
//	     At time.Time
//	 }
//
//	type FooDTO struct {
//	    At string
//	}
//
//	func (f *FooDTO) ToDomain() (*Foo, error) {
//	    var errs []error
//	    out := &Foo{
//	        At: Validate(".at", f.At, &errs, parseTimestamp),
//	    }
//
//	    if len(errs) != 0 {
//	        return nil, errors.Join(errs...)
//	    }
//
//	    return out, nil
//	}
func Validate[I any, O any](fieldName string, in I, errs *[]error, parse func(I) (O, error)) O {
	v, err := parse(in)
	if err != nil {
		var ife *InvalidFieldError
		if errors.As(err, &ife) {
			ife = ife.Qualify(fieldName)
		} else {
			ife = &InvalidFieldError{
				Field: fieldName,
				Err:   err.Error(),
			}
		}

		*errs = append(*errs, ife)
		var zero O
		return zero
	}

	return v
}

// ValidateOptional calls [Validate], but only if in is not the zero value of I.
//
// If in is the zero value of I, the zero value of O is returned.
func ValidateOptional[I comparable, O any](fieldName string, in I, errs *[]error, parse func(I) (O, error)) O {
	if in == util.Zero[I]() {
		return util.Zero[O]()
	}

	return Validate(fieldName, in, errs, parse)
}

// ValidateSlice calls parseOne for each element of in, and returns all successful results.
// If any element failed to parse, an error will be appended to errs, and will not be returned.
//
// This signature makes it easy to map between types in a nice manner, while recording errors.
// For example:
//
//	 type Foo struct {
//	     At []time.Time
//	 }
//
//	type FooDTO struct {
//	    At []string
//	}
//
//	func (f *FooDTO) ToDomain() (*Foo, error) {
//	    var errs []error
//	    out := &Foo{
//	        At: ValidateSlice(".at", f.At, &errs, parseTimestamp),
//	    }
//
//	    if len(errs) != 0 {
//	        return nil, errors.Join(errs...)
//	    }
//
//	    return out, nil
//	}
func ValidateSlice[S ~[]I, I any, O any](fieldName string, in S, errs *[]error, parseOne func(I) (O, error)) []O {
	out := make([]O, 0, len(in))

	for i, v := range in {
		o, err := parseOne(v)
		if err != nil {
			*errs = append(*errs, &InvalidFieldError{
				Field: fieldName + "[" + strconv.Itoa(i) + "]",
				Err:   err.Error(),
			})
			continue
		}

		out = append(out, o)
	}

	return out
}

func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
