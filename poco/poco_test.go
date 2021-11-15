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

		t.Run("nil context", func(t *testing.T) {
			s := serve("testdata/sample.zip")
			c, _ := New(RefreshLocation(s.URL))
			_, err := c.common.client.do(nil, "GET", "/tests", nil)
			assert.NotNil(t, err)
		})
	})

	s := serve("testdata/sample.zip")
	t.Run("can create new instance", func(t *testing.T) {
		c, err := New(UserAgent("test"), RefreshLocation(s.URL), RefreshTimeout(time.Second))
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, 27644, int(c.common.filter.Cap()))
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
			_, err := c.Locations.Get(PostalCode("US", "999999"))
			assert.NotNil(t, err)
		})

		t.Run("db miss", func(t *testing.T) {
			c.common.filter.Add(PostalCode("US", "999999").Bytes())
			_, err := c.Locations.Get(PostalCode("US", "999999"))
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve location", func(t *testing.T) {
		s := serve("testdata/sample.zip")
		c, _ := New(RefreshLocation(s.URL))
		loc, err := c.Locations.Get(PostalCode("US", "20017"))
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, "us", loc.Country)
		assert.Equal(t, "20017", loc.PostalCode)
		assert.Equal(t, 38.9367, loc.Latitude)
		assert.Equal(t, -76.994, loc.Longitude)
		assert.Equal(t, "washington", loc.City)
		assert.Equal(t, "dc", loc.State)
		assert.Equal(t, "district of columbia", loc.County)
	})
}

func TestLocationService_Envelope(t *testing.T) {
	s := serve("testdata/sample.zip")
	c, _ := New(RefreshLocation(s.URL))

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad location", func(t *testing.T) {
			_, err := c.Locations.Envelope(PostalCode("US", "999999"))
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve envelope", func(t *testing.T) {
		env, err := c.Locations.Envelope(City("US", "dc", "washington"))
		assert.Nil(t, err)
		assert.NotNil(t, env)
		assert.Len(t, env.Locations(), 271)
		assert.Equal(t, 38.90095, env.Center().Latitude)
		assert.Equal(t, -77.0118, env.Center().Longitude)
		assert.Equal(t, 10.505164843148291, env.Radius())

		env, err = c.Locations.Envelope(City("US", "ny", "brooklyn"))
		assert.Nil(t, err)
		assert.NotNil(t, env)
		assert.Len(t, env.Locations(), 47)
		assert.Equal(t, 40.65195000000001, env.Center().Latitude)
		assert.Equal(t, -73.95195, env.Center().Longitude)
		assert.Equal(t, 10.665443054330021, env.Radius())

		env, err = c.Locations.Envelope(Country("US"))
		assert.Nil(t, err)
		assert.NotNil(t, env)

		env, err = c.Locations.Envelope(County("US", "ny", "kings"))
		assert.Nil(t, err)
		assert.NotNil(t, env)

		env, err = c.Locations.Envelope(State("US", "ny"))
		assert.Nil(t, err)
		assert.NotNil(t, env)

		env, err = c.Locations.Envelope(State("US", "ny"))
		assert.Nil(t, err)
		assert.NotNil(t, env)
	})

}

func TestParseLocation(t *testing.T) {
	t.Run("notifies on bad location", func(t *testing.T) {
		id, err := ParseLocation("")
		assert.NotNil(t, err)
		assert.Equal(t, "", id.String())
	})

	t.Run("notifies on success", func(t *testing.T) {
		id, err := ParseLocation("country:us")
		assert.Nil(t, err)
		assert.Equal(t, Country("us").String(), id.String())

		id, err = ParseLocation("state:us,ny")
		assert.Nil(t, err)
		assert.Equal(t, State("us", "ny").String(), id.String())

		id, err = ParseLocation("postal:us,20017")
		assert.Nil(t, err)
		assert.Equal(t, PostalCode("us", "20017").String(), id.String())

		id, err = ParseLocation("county:us,ny,kings")
		assert.Nil(t, err)
		assert.Equal(t, County("us", "ny", "kings").String(), id.String())

		id, err = ParseLocation("city:us,ny,brooklyn")
		assert.Nil(t, err)
		assert.Equal(t, City("us", "ny", "brooklyn").String(), id.String())
	})
}

func serve(path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
