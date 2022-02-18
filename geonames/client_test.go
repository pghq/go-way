package geonames

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDB_Get(t *testing.T) {
	t.Parallel()

	t.Run("not ready", func(t *testing.T) {
		var c *Client
		_, err := c.Get(LocationId{})
		assert.NotNil(t, err)
	})

	t.Run("bad location", func(t *testing.T) {
		_, err := NewClient(context.TODO(), "../testdata/sample.zip")
		assert.NotNil(t, err)
	})

	t.Run("bad body", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bad body"))
		}))

		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("too many files", func(t *testing.T) {
		s := serve("../testdata/too-many-files.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("bad columns", func(t *testing.T) {
		s := serve("../testdata/bad-columns.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("bad latitude", func(t *testing.T) {
		s := serve("../testdata/bad-latitude.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("bad longitude", func(t *testing.T) {
		s := serve("../testdata/bad-longitude.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	t.Run("missing index", func(t *testing.T) {
		s := serve("../testdata/missing-index.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.Nil(t, err)
	})

	t.Run("bad quote", func(t *testing.T) {
		s := serve("../testdata/bad-quote.zip")
		_, err := NewClient(context.TODO(), s.URL)
		assert.NotNil(t, err)
	})

	s := serve("../testdata/sample.zip")
	c, _ := NewClient(context.TODO(), s.URL)

	t.Run("with countries", func(t *testing.T) {
		_, err := NewClient(context.TODO(), s.URL, "us")
		assert.Nil(t, err)
	})

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad location", func(t *testing.T) {
			_, err := c.Get(LocationId{})
			assert.NotNil(t, err)

			loc := Location{}
			assert.Equal(t, 0.0, loc.Radius())
			assert.Equal(t, 0.0, loc.Center().Latitude)
			assert.Equal(t, 0.0, loc.Center().Longitude)
		})

		t.Run("not found postal", func(t *testing.T) {
			_, err := c.Get(PostalCode("US", "999999"))
			assert.NotNil(t, err)
		})

		t.Run("not found country", func(t *testing.T) {
			_, err := c.Get(Country("USA"))
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve envelope", func(t *testing.T) {
		loc, err := c.Get(City("US", "dc", "washington"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		assert.Equal(t, 38.90095, loc.Center().Latitude)
		assert.Equal(t, -77.0118, loc.Center().Longitude)
		assert.Equal(t, 10.505164843148291, loc.Radius())

		loc, err = c.Get(City("US", "ny", "brooklyn"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, 40.65195000000001, loc.Center().Latitude)
		assert.Equal(t, -73.95195, loc.Center().Longitude)
		assert.Equal(t, 10.665443054330021, loc.Radius())

		loc, err = c.Get(Country("US"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Get(Secondary("US", "ny", "kings"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Get(Primary("US", "ny"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Get(PostalCode("US", "11201"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)
	})
}

func serve(path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
