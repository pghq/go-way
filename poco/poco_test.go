package poco

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad refresh", func(t *testing.T) {
			_, err := New(RefreshLocation("testdata/sample.zip"))
			assert.NotNil(t, err)
		})

		t.Run("bad response code", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		s := serve("testdata/sample.zip")
		c, _ := New(RefreshLocation(s.URL))

		t.Run("nil client context", func(t *testing.T) {
			_, err := c.common.client.do(nil, "GET", "/tests", nil)
			assert.NotNil(t, err)
		})
	})

	s := serve("testdata/sample.zip")
	t.Run("can create new instance", func(t *testing.T) {
		c, err := New(UserAgent("test"), RefreshLocation(s.URL), RefreshTimeout(time.Second))
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, 95851, int(c.common.cache.Cap()))
		assert.Equal(t, c.common.conf.userAgent, "test")
		assert.Equal(t, c.common.conf.refreshTimeout, time.Second)
		assert.Equal(t, c.common.conf.refreshLocation, s.URL)
	})
}

func TestLocationService_Refresh(t *testing.T) {
	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad body", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bad body"))
			}))

			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("too many files", func(t *testing.T) {
			s := serve("testdata/too-many-files.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("bad columns", func(t *testing.T) {
			s := serve("testdata/bad-columns.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("bad latitude", func(t *testing.T) {
			s := serve("testdata/bad-latitude.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("bad longitude", func(t *testing.T) {
			s := serve("testdata/bad-longitude.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("missing index", func(t *testing.T) {
			s := serve("testdata/missing-index.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})

		t.Run("bad quote", func(t *testing.T) {
			s := serve("testdata/bad-quote.zip")
			_, err := New(RefreshLocation(s.URL))
			assert.NotNil(t, err)
		})
	})
}

func TestLocationService_Export(t *testing.T) {
	s := serve("testdata/sample.zip")
	c, _ := New(RefreshLocation(s.URL))

	t.Run("can create export", func(t *testing.T) {
		r, err := c.Locations.Export()
		assert.Nil(t, err)
		assert.NotNil(t, r)
	})
}

func TestLocationService_Get(t *testing.T) {
	s := serve("testdata/sample.zip")
	c, _ := New(RefreshLocation(s.URL))

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("cache miss", func(t *testing.T) {
			_, err := c.Locations.Get("US", "999999")
			assert.NotNil(t, err)
		})

		t.Run("db miss", func(t *testing.T) {
			c.common.cache.Add(NewLocationId("US", "999999").Bytes())
			_, err := c.Locations.Get("US", "999999")
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve location", func(t *testing.T) {
		loc, err := c.Locations.Get("AR", "4722")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, loc.Country, "AR")
		assert.Equal(t, loc.PostalCode, "4722")
		assert.Equal(t, loc.Latitude, -28.05)
		assert.Equal(t, loc.Longitude, -65.5167)
		assert.Equal(t, loc.AdministrativeName, "Catamarca")
		assert.Equal(t, loc.PlaceName, "SUMAMPA")
	})
}

func serve(path string) *httptest.Server{
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
