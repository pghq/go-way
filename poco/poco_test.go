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
	})
}

func TestLocationService_Export(t *testing.T) {
	s := serve("testdata/sample.zip")
	t.Run("can create export", func(t *testing.T) {
		c, _ := New(RefreshLocation(s.URL))
		r, err := c.Locations.Export()
		assert.Nil(t, err)
		assert.NotNil(t, r)
	})
}

func serve(path string) *httptest.Server{
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
