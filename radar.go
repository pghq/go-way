package way

import (
	"context"
	"sync"
	"time"

	"github.com/pghq/go-way/gndb"
	"github.com/pghq/go-way/mmdb"
)

const (
	// DefaultGeonamesLocation is the default origin location for the GeoName export
	DefaultGeonamesLocation = "https://download.geonames.org/export/zip/allCountries.zip"

	// DefaultMaxmindLocation is the default origin location for the Maxmind export
	DefaultMaxmindLocation = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_LICENSE_KEY&suffix=tar.gz"
)

// Radar is a postal level geo-lookup service.
type Radar struct {
	mutex      sync.RWMutex
	conf       Config
	errors     chan error
	refreshing bool
	gdb        *gndb.DB
	mdb        *mmdb.DB
}

// Wait for refresh
func (r *Radar) Wait() {
	for r.IsRefreshing() {
		<-time.After(time.Microsecond)
	}
}

// IsRefreshing check
func (r *Radar) IsRefreshing() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.refreshing
}

// setRefreshing value
func (r *Radar) setRefreshing(refreshing bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.refreshing = refreshing
}

// Error gets any background errors
func (r *Radar) Error() error {
	select {
	case err := <-r.errors:
		return err
	default:
		return nil
	}
}

// sendError sends backgrounds errors
func (r *Radar) sendError(err error) {
	select {
	case r.errors <- err:
	default:
	}
}

// NewRadar creates a new radar instance.
func NewRadar(opts ...Option) *Radar {
	r := Radar{
		conf: Config{
			refreshTimeout:   DefaultRefreshTimeout,
			geonamesLocation: DefaultGeonamesLocation,
			maxmindLocation:  DefaultMaxmindLocation,
		},
		errors:     make(chan error, 1),
		refreshing: true,
	}
	for _, opt := range opts {
		opt.Apply(&r.conf)
	}

	go r.Refresh(context.Background())

	return &r
}
