package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		_, err := Get(nil, "/tests")
		assert.NotNil(t, err)
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 0)
		defer cancel()
		_, err := Get(ctx, "/tests")
		assert.NotNil(t, err)
	})

	t.Run("bad response code", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

		_, err := Get(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("success", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		_, err := Get(context.TODO(), s.URL)
		assert.Nil(t, err)
	})
}
