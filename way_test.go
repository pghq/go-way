package way

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/pghq/go-tea"
	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-way/country"
)

func TestMain(m *testing.M) {
	tea.Testing()
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	t.Parallel()

	s := serve("testdata/sample.zip")

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad geonames refresh", func(t *testing.T) {
			r := New(GeonamesLocation("testdata/sample.zip"))
			r.Wait()
			assert.NotNil(t, r.Error())
		})

		t.Run("bad maxmind refresh", func(t *testing.T) {
			r := New(GeonamesLocation(s.URL), MaxmindLocation("testdata/GeoLite2-City.tgz"))
			r.Wait()
			assert.NotNil(t, r.Error())
		})

		t.Run("bad response code", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))

			r := New(GeonamesLocation(s.URL))
			r.Wait()
			assert.NotNil(t, r.Error())
		})
	})

	t.Run("can send background errors", func(t *testing.T) {
		r := New(GeonamesLocation(s.URL))
		r.sendError(tea.Err("an error has occurred"))
		r.sendError(tea.Err("an error has occurred"))
	})

	t.Run("can create new instance", func(t *testing.T) {
		r := New(GeonamesLocation(s.URL), RefreshTimeout(time.Second), Countries("us"))
		r.Wait()
		assert.Nil(t, r.Error())
		assert.NotNil(t, r)
		assert.Equal(t, r.refreshTimeout, time.Second)
		assert.Equal(t, r.geonamesLocation, s.URL)
	})
}

func TestRadar_Refresh(t *testing.T) {
	t.Parallel()

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad body", func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("bad body"))
			}))

			r := New(GeonamesLocation(s.URL))
			r.Wait()
			assert.NotNil(t, r.Error())
		})
	})

	t.Run("can refresh", func(t *testing.T) {
		s := serve("testdata/sample.zip")
		mxm := serve("testdata/GeoLite2-City.tgz")
		r := New(GeonamesLocation(s.URL), MaxmindLocation(mxm.URL))
		r.Wait()
		r.Refresh(context.Background())
		r.Wait()
	})
}

func TestRadar_Get(t *testing.T) {
	t.Parallel()

	s := serve("testdata/sample.zip")
	mxm := serve("testdata/GeoLite2-City.tgz")
	r := New(GeonamesLocation(s.URL), MaxmindLocation(mxm.URL), MaxmindKey("test-key"))
	r.Wait()

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("cache miss", func(t *testing.T) {
			_, err := r.Postal("US", "999999")
			assert.NotNil(t, err)
		})

		t.Run("bad ip", func(t *testing.T) {
			_, err := r.IP("bad")
			assert.NotNil(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			_, err := r.IP("192.168.1.1")
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve location", func(t *testing.T) {
		loc, err := r.Postal("US", "20017")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, country.UnitedStatesAmerica, loc.Country)
		assert.Equal(t, "20017", loc.PostalCode)
		assert.Equal(t, 38.9367, loc.Latitude)
		assert.Equal(t, -76.994, loc.Longitude)
		assert.Equal(t, "washington", loc.City)
		assert.Equal(t, "dc", loc.Subdivision1)
		assert.Equal(t, "district of columbia", loc.Subdivision2)

		t.Run("by ip", func(t *testing.T) {
			loc, err := r.IP("81.2.69.142")
			assert.Nil(t, err)
			assert.NotNil(t, loc)

			loc, err = r.IP("216.160.83.56")
			assert.Nil(t, err)
			assert.NotNil(t, loc)
		})
	})
}

func TestRadar_Envelope(t *testing.T) {
	t.Parallel()

	s := serve("testdata/sample.zip")
	r := New(GeonamesLocation(s.URL))
	r.Wait()

	t.Run("should notify on errors", func(t *testing.T) {
		t.Run("bad location", func(t *testing.T) {
			_, err := r.Postal("US", "999999")
			assert.NotNil(t, err)
		})
	})

	t.Run("can retrieve envelope", func(t *testing.T) {
		loc, err := r.City("US", "dc", "washington")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, 38.90095, loc.Center().Latitude)
		assert.Equal(t, -77.0118, loc.Center().Longitude)
		assert.Equal(t, 10.505164843148291, loc.Radius())

		loc, err = r.City("US", "ny", "brooklyn")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
		assert.Equal(t, 40.65195000000001, loc.Center().Latitude)
		assert.Equal(t, -73.95195, loc.Center().Longitude)
		assert.Equal(t, 10.665443054330021, loc.Radius())

		loc, err = r.Country("US")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = r.SSD("US", "ny", "kings")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = r.PSD("US", "ny")
		assert.Nil(t, err)
		assert.NotNil(t, loc)

		loc, err = r.PSD("US", "ny")
		assert.Nil(t, err)
		assert.NotNil(t, loc)
	})
}

func serve(path string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}))
}
