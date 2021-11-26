package way

import (
	"context"
	"strings"
	"time"

	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/geonames"
	"github.com/pghq/go-way/maxmind"
)

const (
	// DefaultRefreshTimeout is the default wait time for refreshing postal codes from GeoNames
	DefaultRefreshTimeout = 60 * time.Second
)

// Refresh postal codes from GeoNames and cache
func (s *LocationService) Refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.conf.refreshTimeout)
	defer cancel()

	gdb, err := geonames.Open(ctx, s.conf.geonamesLocation)
	if err != nil {
		return tea.Error(err)
	}
	s.gdb = gdb

	if s.conf.maxmindLocation != DefaultMaxmindLocation || s.conf.maxmindKey != "" {
		mdb, err := maxmind.Open(ctx, strings.Replace(s.conf.maxmindLocation, "YOUR_LICENSE_KEY", s.conf.maxmindKey, 1))
		if err != nil {
			return tea.Error(err)
		}

		if s.mdb != nil {
			_ = s.mdb.Close()
		}

		s.mdb = mdb
	}

	return nil
}
