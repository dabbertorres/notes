package apiv1

import (
	"fmt"
	"net/http"
	"time"
)

type InvalidFieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

func (e *InvalidFieldError) Status() int   { return http.StatusBadRequest }
func (e *InvalidFieldError) Body() any     { return e }
func (e *InvalidFieldError) Error() string { return fmt.Sprintf("%s: %v", e.Field, e.Err) }

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
		*errs = append(*errs, &InvalidFieldError{
			Field: fieldName,
			Err:   err.Error(),
		})
		var zero O
		return zero
	}

	return v
}

func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
