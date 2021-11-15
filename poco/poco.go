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

// Package poco provides a client for postal code lookups.
package poco

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/hashicorp/go-memdb"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// DefaultUserAgent is the default user agent for outgoing requests
	DefaultUserAgent = "go-poco/v0"
)

// Client allows interaction with services within the domain.
type Client struct {
	common service
	errors chan error

	Locations *LocationService
}

// New creates a new client instance.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		errors: make(chan error, 1),
	}
	c.common.conf = Config{
		refreshTimeout:  DefaultRefreshTimeout,
		refreshLocation: DefaultRefreshLocation,
		userAgent:       DefaultUserAgent,
	}

	for _, opt := range opts {
		opt.Apply(&c.common.conf)
	}

	c.common.client.userAgent = c.common.conf.userAgent
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"locations": {
				Name: "locations",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Country"},
								&memdb.StringFieldIndex{Field: "PostalCode"},
							},
						},
					},
					"country": {
						Name:    "country",
						Indexer: &memdb.StringFieldIndex{Field: "Country"},
					},
					"state": {
						Name:         "state",
						AllowMissing: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Country"},
								&memdb.StringFieldIndex{Field: "State"},
							},
						},
					},
					"county": {
						Name:         "county",
						AllowMissing: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Country"},
								&memdb.StringFieldIndex{Field: "State"},
								&memdb.StringFieldIndex{Field: "County"},
							},
						},
					},
					"city": {
						Name:         "city",
						AllowMissing: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Country"},
								&memdb.StringFieldIndex{Field: "State"},
								&memdb.StringFieldIndex{Field: "City"},
							},
						},
					},
				},
			},
		},
	}

	// at this point we have a valid schema
	c.common.db, _ = memdb.NewMemDB(schema)

	c.Locations = (*LocationService)(&c.common)
	if err := c.Locations.Refresh(context.Background()); err != nil {
		return nil, errors.Wrap(err)
	}

	return c, nil
}

// Config for the client
type Config struct {
	userAgent       string
	refreshLocation string
	refreshTimeout  time.Duration
}

// Option for configuring the client
type Option interface {
	Apply(conf *Config)
}

type userAgent string

func (o userAgent) Apply(conf *Config) {
	if conf != nil {
		conf.userAgent = string(o)
	}
}

// UserAgent is an option for configuring the user agent
func UserAgent(agent string) Option {
	return userAgent(agent)
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

type refreshLocation string

func (o refreshLocation) Apply(conf *Config) {
	if conf != nil {
		conf.refreshLocation = string(o)
	}
}

// RefreshLocation is an option for configuring the refresh location
func RefreshLocation(origin string) Option {
	return refreshLocation(origin)
}

// service is a shared configuration for all services within the domain.
type service struct {
	conf   Config
	client client
	filter *bloom.BloomFilter
	db     *memdb.MemDB
	source *bytes.Reader
}

// client is a shared http client for services
type client struct {
	userAgent string
}

// Get http request
func (c *client) Get(ctx context.Context, url string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, url, nil)
}

// do a http request
func (c *client) do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	r.Header.Set("User-Agent", c.userAgent)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if resp.StatusCode != 200 {
		return nil, errors.Newf("unexpected refresh response code %d", resp.StatusCode)
	}

	return resp, nil
}
