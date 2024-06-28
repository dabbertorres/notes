package apiv1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageMarshalJSON(t *testing.T) {
	t.Run("has_next_page", func(t *testing.T) {
		page := Page[int]{
			NextPageToken: &NextPageToken{
				NextItemID: "131",
				PageSize:   100,
			},
			Items: []int{5, 3, 1},
		}

		data, err := json.Marshal(&page)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"next_page_token": "MTMxOzEwMA", "items": [5, 3, 1]}`, string(data))
	})

	t.Run("is_last_page", func(t *testing.T) {
		page := Page[int]{
			NextPageToken: nil,
			Items:         []int{5, 3, 1},
		}

		data, err := json.Marshal(&page)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"items": [5, 3, 1]}`, string(data))
	})
}

func TestPageUnmarshalJSON(t *testing.T) {
	t.Run("has_next_page", func(t *testing.T) {
		data := `{"next_page_token": "MTMxOzEwMA", "items": [5, 3, 1]}`

		var page Page[int]
		err := json.Unmarshal([]byte(data), &page)
		assert.NoError(t, err)

		expect := Page[int]{
			NextPageToken: &NextPageToken{
				NextItemID: "131",
				PageSize:   100,
			},
			Items: []int{5, 3, 1},
		}

		assert.Equal(t, expect, page)
	})

	t.Run("last_page", func(t *testing.T) {
		data := `{"items": [5, 3, 1]}`

		var page Page[int]
		err := json.Unmarshal([]byte(data), &page)
		assert.NoError(t, err)

		expect := Page[int]{
			NextPageToken: nil,
			Items:         []int{5, 3, 1},
		}

		assert.Equal(t, expect, page)
	})
}

func TestNextPageTokenMarshalText(t *testing.T) {
	token := NextPageToken{
		NextItemID: "131",
		PageSize:   100,
	}

	out, err := token.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("MTMxOzEwMA"), out)
}

func TestNextPageTokenUnmarshalText(t *testing.T) {
	text := "MTMxOzEwMA"

	var token NextPageToken
	err := token.UnmarshalText([]byte(text))
	assert.NoError(t, err)

	expect := NextPageToken{
		NextItemID: "131",
		PageSize:   100,
	}
	assert.Equal(t, expect, token)
}

func TestNextPageTokenMarshalRoundTrip(t *testing.T) {
	input := NextPageToken{
		NextItemID: "131",
		PageSize:   100,
	}

	out, err := input.MarshalText()
	assert.NoError(t, err)

	var output NextPageToken
	err = output.UnmarshalText(out)
	assert.NoError(t, err)

	assert.Equal(t, input, output)
}
