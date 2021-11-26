package way

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	s := serve("testdata/sample.zip")

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad geonames refresh", func(t *testing.T) {
			_, err := NewClient(GeonamesLocation("testdata/sample.zip"))
			assert.NotNil(t, err)
		})

		t.Run("bad maxmind refresh", func(t *testing.T) {
			_, err := NewClient(GeonamesLocation(s.URL), MaxmindLocation("testdata/maxmind-test.tgz"))
			assert.NotNil(t, err)
		})

		t.Run("bad response code", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			_, err := NewClient(GeonamesLocation(s.URL))
			assert.NotNil(t, err)
		})
	})

	t.Run("can create new instance", func(t *testing.T) {
		c, err := NewClient(GeonamesLocation(s.URL), RefreshTimeout(time.Second))
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, c.common.conf.refreshTimeout, time.Second)
		assert.Equal(t, c.common.conf.geonamesLocation, s.URL)
	})
}

func TestLocationService_Refresh(t *testing.T) {
	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad body", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bad body"))
			}))

			_, err := NewClient(GeonamesLocation(s.URL))
			assert.NotNil(t, err)
		})
	})

	s := serve("testdata/sample.zip")
	mxm := serve("testdata/maxmind-test.tgz")
	c, _ := NewClient(GeonamesLocation(s.URL), MaxmindLocation(mxm.URL))

	t.Run("can refresh", func(t *testing.T) {
		assert.Nil(t, c.Locations.Refresh(context.TODO()))
	})
}

func TestLocationService_Get(t *testing.T) {
	s := serve("testdata/sample.zip")
	mxm := serve("testdata/maxmind-test.tgz")
	c, _ := NewClient(GeonamesLocation(s.URL), MaxmindLocation(mxm.URL), MaxmindKey("test-key"))
	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("cache miss", func(t *testing.T) {
			_, err := c.Locations.Postal("US", "999999")
			assert.NotNil(t, err)
		})

		t.Run("bad ip", func(t *testing.T) {
			_, err := c.Locations.IP("bad")
			assert.NotNil(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			_, err := c.Locations.IP("192.168.1.1")
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve location", func(t *testing.T) {
		loc, err := c.Locations.Postal("US", "20017")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, "us", loc.Country)
		assert.Equal(t, "20017", loc.PostalCode)
		assert.Equal(t, 38.9367, loc.Latitude)
		assert.Equal(t, -76.994, loc.Longitude)
		assert.Equal(t, "washington", loc.City)
		assert.Equal(t, "dc", loc.Subdivision1)
		assert.Equal(t, "district of columbia", loc.Subdivision2)

		t.Run("by ip", func(t *testing.T) {
			loc, err := c.Locations.IP("81.2.69.142")
			assert.Nil(t, err)
			assert.NotNil(t, loc)

			loc, err = c.Locations.IP("216.160.83.56")
			assert.Nil(t, err)
			assert.NotNil(t, loc)
		})
	})
}

func TestLocationService_Envelope(t *testing.T) {
	s := serve("testdata/sample.zip")
	c, _ := NewClient(GeonamesLocation(s.URL))

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad location", func(t *testing.T) {
			_, err := c.Locations.Postal("US", "999999")
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve envelope", func(t *testing.T) {
		loc, err := c.Locations.City("US", "dc", "washington")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, 38.90095, loc.Center().Latitude)
		assert.Equal(t, -77.0118, loc.Center().Longitude)
		assert.Equal(t, 10.505164843148291, loc.Radius())

		loc, err = c.Locations.City("US", "ny", "brooklyn")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, 40.65195000000001, loc.Center().Latitude)
		assert.Equal(t, -73.95195, loc.Center().Longitude)
		assert.Equal(t, 10.665443054330021, loc.Radius())

		loc, err = c.Locations.Country("US")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Locations.Secondary("US", "ny", "kings")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Locations.Primary("US", "ny")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = c.Locations.Primary("US", "ny")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
	})
}

func serve(path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
