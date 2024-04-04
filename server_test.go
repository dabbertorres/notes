package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_chainMiddleware(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var handlerCalled bool

		mw := chainMiddleware()
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}))

		assert.NotPanics(t, func() { h.ServeHTTP(nil, nil) })
		assert.True(t, handlerCalled, "handlerCalled")
	})

	t.Run("one", func(t *testing.T) {
		var (
			middlewareCalled bool
			handlerCalled    bool
		)
		mw := chainMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		})

		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}))

		assert.NotPanics(t, func() { h.ServeHTTP(nil, nil) })
		assert.True(t, middlewareCalled, "middlewareCalled")
		assert.True(t, handlerCalled, "handlerCalled")
	})

	t.Run("multiple in expected order", func(t *testing.T) {
		var (
			oneCalled     int
			twoCalled     int
			threeCalled   int
			handlerCalled int

			hits int
		)

		recordHit := func(i *int) {
			hits++
			*i = hits
		}

		recordMW := func(i *int) middleware {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					recordHit(i)
					next.ServeHTTP(w, r)
				})
			}
		}

		mw := chainMiddleware(
			recordMW(&oneCalled),
			recordMW(&twoCalled),
			recordMW(&threeCalled),
		)

		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recordHit(&handlerCalled)
		}))

		assert.NotPanics(t, func() { h.ServeHTTP(nil, nil) })

		assert.Equal(t, 1, oneCalled, "oneCalled")
		assert.Equal(t, 2, twoCalled, "twoCalled")
		assert.Equal(t, 3, threeCalled, "threeCalled")
		assert.Equal(t, 4, handlerCalled, "handlerCalled")
	})
}
