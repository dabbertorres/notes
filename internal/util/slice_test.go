package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceDiff(t *testing.T) {
	type testCase struct {
		name   string
		before []int
		after  []int

		wantAdditions []int
		wantDeletions []int
	}

	cases := []testCase{
		{
			name:   "equal",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 3, 4, 5},

			wantAdditions: nil,
			wantDeletions: nil,
		},
		{
			name:   "all adds",
			before: nil,
			after:  []int{1, 2, 3, 4, 5},

			wantAdditions: []int{1, 2, 3, 4, 5},
			wantDeletions: nil,
		},
		{
			name:   "all deletes",
			before: []int{1, 2, 3, 4, 5},
			after:  nil,

			wantAdditions: nil,
			wantDeletions: []int{1, 2, 3, 4, 5},
		},
		{
			name:   "completely different",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{6, 7, 8, 9, 10},

			wantAdditions: []int{6, 7, 8, 9, 10},
			wantDeletions: []int{1, 2, 3, 4, 5},
		},
		{
			name:   "addition/beginning",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{0, 1, 2, 3, 4, 5},

			wantAdditions: []int{0},
			wantDeletions: nil,
		},
		{
			name:   "addition/ending",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 3, 4, 5, 6},

			wantAdditions: []int{6},
			wantDeletions: nil,
		},
		{
			name:   "deletion/beginning",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{2, 3, 4, 5},

			wantAdditions: nil,
			wantDeletions: []int{1},
		},
		{
			name:   "deletion/ending",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 3, 4},

			wantAdditions: nil,
			wantDeletions: []int{5},
		},
		{
			name:   "deletion/center",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 4, 5},

			wantAdditions: nil,
			wantDeletions: []int{3},
		},
		{
			name:   "deletion/multiple",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 5},

			wantAdditions: nil,
			wantDeletions: []int{2, 3, 4},
		},
		{
			name:   "different/beginning",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{0, 2, 3, 4, 5},

			wantAdditions: []int{0},
			wantDeletions: []int{1},
		},
		{
			name:   "different/ending",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 3, 4, 6},

			wantAdditions: []int{6},
			wantDeletions: []int{5},
		},
		{
			name:   "different/center",
			before: []int{1, 2, 3, 5, 6},
			after:  []int{1, 2, 4, 5, 6},

			wantAdditions: []int{4},
			wantDeletions: []int{3},
		},
		{
			name:   "duplicates",
			before: []int{1, 2, 3, 4, 5},
			after:  []int{1, 2, 3, 3, 5},

			wantAdditions: []int{3},
			wantDeletions: []int{4},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualAdditions, actualDeletions := SliceDiff(tc.before, tc.after)

			assert.Equal(t, tc.wantAdditions, actualAdditions, "additions")
			assert.Equal(t, tc.wantDeletions, actualDeletions, "deletions")
		})
	}
}
