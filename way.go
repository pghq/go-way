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
	"time"

	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/geonames"
	"github.com/pghq/go-way/maxmind"
)

const (
	// DefaultGeonamesLocation is the default origin location for the GeoName export
	DefaultGeonamesLocation = "https://download.geonames.org/export/zip/allCountries.zip"

	// DefaultMaxmindLocation is the default origin location for the Maxmind export
	DefaultMaxmindLocation = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=YOUR_LICENSE_KEY&suffix=tar.gz"
)

// Client allows interaction with services within the domain.
type Client struct {
	common service

	Locations *LocationService
}

// NewClient creates a new client instance.
func NewClient(opts ...Option) (*Client, error) {
	c := Client{}
	c.common.conf = Config{
		refreshTimeout:   DefaultRefreshTimeout,
		geonamesLocation: DefaultGeonamesLocation,
		maxmindLocation:  DefaultMaxmindLocation,
	}

	for _, opt := range opts {
		opt.Apply(&c.common.conf)
	}

	c.Locations = (*LocationService)(&c.common)
	if err := c.Locations.Refresh(context.Background()); err != nil {
		return nil, tea.Error(err)
	}

	return &c, nil
}

// Config for the client
type Config struct {
	userAgent        string
	geonamesLocation string
	maxmindLocation  string
	maxmindKey       string
	refreshTimeout   time.Duration
}

// Option for configuring the client
type Option interface {
	Apply(conf *Config)
}

type refreshTimeout time.Duration

func (o refreshTimeout) Apply(conf *Config) {
	if conf != nil {
		conf.refreshTimeout = time.Duration(o)
	}
}

// RefreshTimeout is an option for configuring the refresh timeout
func RefreshTimeout(t time.Duration) Option {
	return refreshTimeout(t)
}

type geonamesLocation string

func (o geonamesLocation) Apply(conf *Config) {
	if conf != nil {
		conf.geonamesLocation = string(o)
	}
}

// GeonamesLocation is an option for configuring the geonames location
func GeonamesLocation(origin string) Option {
	return geonamesLocation(origin)
}

type maxmindLocation string

func (o maxmindLocation) Apply(conf *Config) {
	if conf != nil {
		conf.maxmindLocation = string(o)
	}
}

// MaxmindLocation is an option for configuring the maxmind location
func MaxmindLocation(origin string) Option {
	return maxmindLocation(origin)
}

type maxmindKey string

func (o maxmindKey) Apply(conf *Config) {
	if conf != nil {
		conf.maxmindKey = string(o)
	}
}

// MaxmindKey configures the maxmind licence key
func MaxmindKey(key string) Option {
	return maxmindKey(key)
}

// service is a shared configuration for all services within the domain.
type service struct {
	conf Config
	gdb  *geonames.DB
	mdb  *maxmind.DB
}
