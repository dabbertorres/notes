package apiv1

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

type Pager interface {
	EncodePager() ([][]byte, error)
	DecodePager([][]byte) error
}

type Page[T any, P Pager] struct {
	NextPageToken *PageToken[P] `json:"next_page_token,omitempty"`
	Items         []T           `json:"items,omitempty"`
}

type PageToken[P Pager] struct {
	Data     P
	PageSize int
}

func ParsePageToken[P Pager](s string, defaultPageSize, maxPageSize int) (token PageToken[P], err error) {
	token.PageSize = defaultPageSize
	if len(s) == 0 {
		return
	}

	token.Data = reflect.New(reflect.TypeFor[P]().Elem()).Interface().(P)
	if err = token.UnmarshalText([]byte(s)); err != nil {
		err = NewError(http.StatusBadRequest, "invalid page token")
		return
	}

	if token.PageSize > maxPageSize {
		err = NewError(http.StatusBadRequest, "requested page size is too large")
		return
	}

	return
}

var pageTokenEncoding = base64.RawURLEncoding

func (t *PageToken[P]) MarshalText() ([]byte, error) {
	if t == nil {
		return nil, nil
	}

	fields, err := t.Data.EncodePager()
	if err != nil {
		return nil, err
	}

	unencoded := bytes.Join(fields, []byte{';'})
	unencoded = append(unencoded, ';')
	unencoded = strconv.AppendInt(unencoded, int64(t.PageSize), 10)

	out := make([]byte, pageTokenEncoding.EncodedLen(len(unencoded)))
	pageTokenEncoding.Encode(out, unencoded)
	return out, nil
}

func (t *PageToken[P]) UnmarshalText(data []byte) (err error) {
	unencoded := make([]byte, pageTokenEncoding.DecodedLen(len(data)))
	if _, err := pageTokenEncoding.Decode(unencoded, data); err != nil {
		return err
	}

	parts := bytes.Split(unencoded, []byte{';'})
	// at least page size must be present
	if len(parts) < 1 {
		return errors.New("invalid page token format (incorrect number of parts)")
	}

	t.Data = reflect.New(reflect.TypeFor[P]().Elem()).Interface().(P)
	if err := t.Data.DecodePager(parts[:len(parts)-1]); err != nil {
		return err
	}

	t.PageSize, err = strconv.Atoi(string(parts[len(parts)-1]))
	if err != nil {
		return err
	}

	return nil
}
