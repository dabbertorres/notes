package apiv1

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

type Page[T any] struct {
	NextPageToken *NextPageToken `json:"next_page_token,omitempty"`
	Items         []T            `json:"items"`
}

type NextPageToken struct {
	NextItemID string
	PageSize   int
}

var pageTokenEncoding = base64.RawURLEncoding

func (t *NextPageToken) MarshalText() (out []byte, err error) {
	if t == nil {
		return nil, nil
	}

	unencoded := fmt.Appendf(nil, "%s;%d", t.NextItemID, t.PageSize)

	out = make([]byte, pageTokenEncoding.EncodedLen(len(unencoded)))
	pageTokenEncoding.Encode(out, unencoded)
	return out, nil
}

func (t *NextPageToken) UnmarshalText(data []byte) (err error) {
	unencoded := make([]byte, pageTokenEncoding.DecodedLen(len(data)))
	if _, err := pageTokenEncoding.Decode(unencoded, data); err != nil {
		return err
	}

	parts := bytes.SplitN(unencoded, []byte{';'}, 2)
	if len(parts) != 2 {
		return errors.New("invalid page token format (incorrect number of parts)")
	}

	t.NextItemID = string(parts[0])
	t.PageSize, err = strconv.Atoi(string(parts[1]))
	if err != nil {
		return err
	}

	return nil
}
