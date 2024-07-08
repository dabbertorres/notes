package apiv1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testPageTokenData struct {
	LastItemID string
}

func (t *testPageTokenData) EncodePager() ([][]byte, error) {
	return [][]byte{
		[]byte(t.LastItemID),
	}, nil
}

func (t *testPageTokenData) DecodePager(data [][]byte) error {
	if t == nil {
		return nil
	}

	t.LastItemID = string(data[0])
	return nil
}

func TestPage_MarshalJSON(t *testing.T) {
	t.Run("has_next_page", func(t *testing.T) {
		page := Page[int, *testPageTokenData]{
			NextPageToken: &PageToken[*testPageTokenData]{
				Data:     &testPageTokenData{LastItemID: "131"},
				PageSize: 100,
			},
			Items: []int{5, 3, 1},
		}

		data, err := json.Marshal(&page)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"next_page_token": "MTMxOzEwMA", "items": [5, 3, 1]}`, string(data))
	})

	t.Run("is_last_page", func(t *testing.T) {
		page := Page[int, *testPageTokenData]{
			NextPageToken: nil,
			Items:         []int{5, 3, 1},
		}

		data, err := json.Marshal(&page)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"items": [5, 3, 1]}`, string(data))
	})
}

func TestPage_UnmarshalJSON(t *testing.T) {
	t.Run("has_next_page", func(t *testing.T) {
		data := `{"next_page_token": "MTMxOzEwMA", "items": [5, 3, 1]}`

		var page Page[int, *testPageTokenData]
		err := json.Unmarshal([]byte(data), &page)
		assert.NoError(t, err)

		expect := Page[int, *testPageTokenData]{
			NextPageToken: &PageToken[*testPageTokenData]{
				Data:     &testPageTokenData{LastItemID: "131"},
				PageSize: 100,
			},
			Items: []int{5, 3, 1},
		}

		assert.Equal(t, expect, page)
	})

	t.Run("last_page", func(t *testing.T) {
		data := `{"items": [5, 3, 1]}`

		var page Page[int, *testPageTokenData]
		err := json.Unmarshal([]byte(data), &page)
		assert.NoError(t, err)

		expect := Page[int, *testPageTokenData]{
			NextPageToken: nil,
			Items:         []int{5, 3, 1},
		}

		assert.Equal(t, expect, page)
	})
}

func TestPageToken_MarshalText(t *testing.T) {
	token := PageToken[*testPageTokenData]{
		Data:     &testPageTokenData{LastItemID: "131"},
		PageSize: 100,
	}

	out, err := token.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("MTMxOzEwMA"), out)
}

func TestPageToken_UnmarshalText(t *testing.T) {
	text := "MTMxOzEwMA"

	var token PageToken[*testPageTokenData]
	err := token.UnmarshalText([]byte(text))
	assert.NoError(t, err)

	expect := PageToken[*testPageTokenData]{
		Data:     &testPageTokenData{LastItemID: "131"},
		PageSize: 100,
	}
	assert.Equal(t, expect, token)
}

func TestPageToken_MarshalRoundTrip(t *testing.T) {
	input := PageToken[*testPageTokenData]{
		Data:     &testPageTokenData{LastItemID: "131"},
		PageSize: 100,
	}

	out, err := input.MarshalText()
	assert.NoError(t, err)

	var output PageToken[*testPageTokenData]
	err = output.UnmarshalText(out)
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
