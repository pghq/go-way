// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package way provides a client for postal code lookups.
package way

import (
	"context"
	"sync"
	"time"

	"github.com/pghq/go-red"

	"github.com/pghq/go-way/geonames"
	"github.com/pghq/go-way/maxmind"
)

const (
	// DefaultGeonamesLocation is the default origin location for the GeoName export
	DefaultGeonamesLocation = "https://download.geonames.org/export/zip/allCountries.zip"

	// DefaultMaxmindLocation is the default origin location for the Maxmind export
	DefaultMaxmindLocation = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_LICENSE_KEY&suffix=tar.gz"

	// DefaultRefreshTimeout is the default wait time for refreshing locations
	DefaultRefreshTimeout = 5 * time.Minute
)

// Radar is a postal level geo-lookup service.
type Radar struct {
	userAgent        string
	geonamesLocation string
	maxmindLocation  string
	maxmindKey       string
	refreshTimeout   time.Duration
	errors           chan error
	waits            chan *sync.WaitGroup
	refreshes        chan context.Context
	bg               *red.Worker
	geonames         *geonames.Client
	maxmind          *maxmind.Client
}

// Wait for refresh
func (r *Radar) Wait() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	r.waits <- &wg
	wg.Wait()
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

// New creates a new radar instance.
func New(opts ...RadarOption) *Radar {
	r := Radar{
		refreshTimeout:   DefaultRefreshTimeout,
		geonamesLocation: DefaultGeonamesLocation,
		maxmindLocation:  DefaultMaxmindLocation,
		errors:           make(chan error, 1),
		waits:            make(chan *sync.WaitGroup),
		refreshes:        make(chan context.Context, 1),
	}

	bg := red.NewWorker(r.refreshJob)
	for _, opt := range opts {
		opt(&r)
	}

	r.bg = bg
	go bg.Start()
	go r.Refresh(context.Background())

	return &r
}

// RadarOption to configure custom radar
type RadarOption func(r *Radar)

// RefreshTimeout sets a custom refresh timeout
func RefreshTimeout(o time.Duration) RadarOption {
	return func(r *Radar) {
		r.refreshTimeout = o
	}
}

// GeonamesLocation sets a custom location to refresh geonames db from
func GeonamesLocation(o string) RadarOption {
	return func(r *Radar) {
		r.geonamesLocation = o
	}
}

// MaxmindLocation sets a custom location to refresh maxmind db from
func MaxmindLocation(o string) RadarOption {
	return func(r *Radar) {
		r.maxmindLocation = o
	}
}

// MaxmindKey sets a custom maxmind licence key
func MaxmindKey(o string) RadarOption {
	return func(r *Radar) {
		r.maxmindKey = o
	}
}
