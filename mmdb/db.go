package mmdb

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/pghq/go-ark"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// positiveTTL is the positive ttl for search queries
	positiveTTL = 30 * time.Minute
)

// DB instance for MaxMind
type DB struct {
	Size int
	mdb  *geoip2.Reader
	conn *ark.KVSConn
}

// Get city
func (db *DB) Get(ip net.IP) (*geoip2.City, error) {
	if db == nil {
		return nil, tea.NewNoContent("db not ready")
	}

	var city *geoip2.City
	return city, db.conn.Do(context.Background(), func(tx *ark.KVSTxn) error {
		var cy geoip2.City
		if _, err := tx.Get([]byte(ip.String()), &cy).Resolve(); err == nil {
			city = &cy
			return nil
		}

		c, err := db.mdb.City(ip)
		if err != nil {
			return tea.Error(err)
		}

		if c == nil || c.City.GeoNameID == 0 {
			return tea.NewNoContent("not found")
		}

		city = c
		tx.InsertWithTTL([]byte(ip.String()), city, positiveTTL)
		return nil
	})
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

		base := filepath.Base(header.Name)
		if !strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".mmdb") {
			break
		}
	}

	b, _ = ioutil.ReadAll(tr)
	mdb, err := geoip2.FromBytes(b)
	if err != nil {
		return nil, tea.Error(err)
	}

	db := DB{
		mdb: mdb,
	}

	db.Size = int(db.mdb.Metadata().NodeCount)
	dm := ark.Open()
	db.conn, _ = dm.ConnectKVS(ctx, "inmem")

	return &db, nil
}
