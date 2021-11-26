package maxmind

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/oschwald/geoip2-golang"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// positiveTTL is the positive ttl for search queries
	positiveTTL = 30 * time.Minute

	// negativeTTL is the negative ttl for search queries
	negativeTTL = 90 * time.Minute
)

// DB instance for maxmind
type DB struct {
	mdb   *geoip2.Reader
	cache *ristretto.Cache
}

// Get city
func (db *DB) Get(ip net.IP) (*geoip2.City, error) {
	v, present := db.cache.Get(ip.String())
	if present {
		if err, ok := v.(error); ok {
			return nil, tea.Error(err)
		}

		if city, ok := v.(*geoip2.City); ok {
			return city, nil
		}
	}

	city, err := db.mdb.City(ip)
	if err != nil {
		return nil, tea.Error(err)
	}

	if city == nil || city.City.GeoNameID == 0 {
		err := tea.NewNoContent()
		db.cache.SetWithTTL(ip.String(), err, 1, negativeTTL)
		return nil, err
	}

	db.cache.SetWithTTL(ip.String(), city, 1, positiveTTL)
	return city, nil
}

// Close the db
func (db *DB) Close() error {
	return db.mdb.Close()
}

// Open the maxmind db
func Open(ctx context.Context, uri string) (*DB, error) {
	resp, err := client.Get(ctx, uri)
	if err != nil {
		return nil, tea.Error(err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	stream, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, tea.Error(err)
	}

	tr := tar.NewReader(stream)
	for {
		header, err := tr.Next()
		if err != nil {
			return nil, tea.Error(err)
		}

		if strings.HasPrefix(header.Name, ".") {
			continue
		}

		break
	}

	b, _ = ioutil.ReadAll(tr)
	mdb, err := geoip2.FromBytes(b)
	if err != nil {
		return nil, tea.Error(err)
	}

	db := DB{
		mdb: mdb,
	}
	db.cache, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	return &db, nil
}
