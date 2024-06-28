package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChain1(t *testing.T) {
	var indices []int

	chained := Chain1(
		func(i int) int {
			indices = append(indices, 0)
			return i + 1
		},
		func(i int) int {
			indices = append(indices, 1)
			return i + 1
		},
		func(i int) int {
			indices = append(indices, 2)
			return i + 1
		},
		func(i int) int {
			indices = append(indices, 3)
			return i + 1
		},
		func(i int) int {
			indices = append(indices, 4)
			return i + 1
		},
	)
	out := chained(0)

	assert.Equal(t, 5, out)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, indices)
}

func TestChainReverse1(t *testing.T) {
	var indices []int

	type handler func(int)

	chained := ChainReverse1(
		func(next handler) handler {
			return func(i int) {
				indices = append(indices, 0)
				next(i + 1)
			}
		},
		func(next handler) handler {
			return func(i int) {
				indices = append(indices, 1)
				next(i + 1)
			}
		},
		func(next handler) handler {
			return func(i int) {
				indices = append(indices, 2)
				next(i + 1)
			}
		},
		func(next handler) handler {
			return func(i int) {
				indices = append(indices, 3)
				next(i + 1)
			}
		},
		func(next handler) handler {
			return func(i int) {
				indices = append(indices, 4)
				next(i + 1)
			}
		},
	)

	var final int

	out := chained(func(i int) {
		final = i
	})

	out(0)

	assert.Equal(t, 5, final)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, indices)
}
