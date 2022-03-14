package way

import (
	"context"
	"strings"
	"sync"

	"github.com/pghq/go-way/geonames"
	"github.com/pghq/go-way/maxmind"
)

// Refresh locations
func (r *Radar) Refresh() {
	wg := sync.WaitGroup{}
	r.refreshes <- &wg
	wg.Add(1)
	wg.Wait()
}

// refreshJob process pending refreshes / waits
func (r *Radar) refreshJob() {
	select {
	case wg := <-r.refreshes:
		defer wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), r.refreshTimeout)
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
