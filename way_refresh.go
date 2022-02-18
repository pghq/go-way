package way

import (
	"context"
	"strings"

	"github.com/pghq/go-way/geonames"
	"github.com/pghq/go-way/maxmind"
)

// Refresh locations
func (r *Radar) Refresh(ctx context.Context) {
	r.refreshes <- ctx
}

// refreshJob process pending refreshes / waits
func (r *Radar) refreshJob(_ context.Context) {
	defer func() {
		for {
			select {
			case wg := <-r.waits:
				wg.Done()
			default:
				return
			}
		}
	}()

	select {
	case ctx := <-r.refreshes:
		ctx, cancel := context.WithTimeout(ctx, r.refreshTimeout)
		defer cancel()

		gc, err := geonames.NewClient(ctx, r.geonamesLocation, r.countries...)
		if err != nil {
			r.sendError(err)
			return
		}

		r.geonames = gc
		if r.maxmindLocation != DefaultMaxmindLocation || r.maxmindKey != "" {
			mc, err := maxmind.NewClient(ctx, strings.Replace(r.maxmindLocation, "YOUR_LICENSE_KEY", r.maxmindKey, 1))
			if err != nil {
				r.sendError(err)
				return
			}

			if r.maxmind != nil {
				_ = r.maxmind.Close()
			}

			r.maxmind = mc
		}
	default:
	}
}
