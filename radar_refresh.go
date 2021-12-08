package way

import (
	"context"
	"strings"
	"time"

	"github.com/pghq/go-way/gndb"
	"github.com/pghq/go-way/mmdb"
)

const (
	// DefaultRefreshTimeout is the default wait time for refreshing locations
	DefaultRefreshTimeout = 5 * time.Minute
)

// Refresh locations
func (r *Radar) Refresh(ctx context.Context) {
	r.setRefreshing(true)
	defer r.setRefreshing(false)

	ctx, cancel := context.WithTimeout(ctx, r.conf.refreshTimeout)
	defer cancel()

	gdb, err := gndb.Open(ctx, r.conf.geonamesLocation)
	if err != nil {
		r.sendError(err)
		return
	}

	r.mutex.Lock()
	r.gdb = gdb
	r.mutex.Unlock()

	if r.conf.maxmindLocation != DefaultMaxmindLocation || r.conf.maxmindKey != "" {
		mdb, err := mmdb.Open(ctx, strings.Replace(r.conf.maxmindLocation, "YOUR_LICENSE_KEY", r.conf.maxmindKey, 1))
		if err != nil {
			r.sendError(err)
			return
		}

		r.mutex.Lock()
		if r.mdb != nil {
			_ = r.mdb.Close()
		}

		r.mdb = mdb
		r.mutex.Unlock()
	}
}
