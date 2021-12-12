package maxmind

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
	"github.com/pghq/go-ark/db"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// PositiveTTL is the positive ttl for search queries
	PositiveTTL = 30 * time.Minute
)

// Client for Maxmind
type Client struct {
	IPCount int
	reader  *geoip2.Reader
	mapper  *ark.Mapper
}

// Get city by id
func (c *Client) Get(ip net.IP) (*geoip2.City, error) {
	if c == nil {
		return nil, tea.NewNoContent("client not ready")
	}

	var city *geoip2.City
	return city, c.mapper.Do(context.Background(), func(tx db.Txn) error {
		var cy geoip2.City
		if err := tx.Get("", ip.String(), &cy); err == nil {
			city = &cy
			return nil
		}

		c, err := c.reader.City(ip)
		if err != nil {
			return tea.Error(err)
		}

		if c == nil || c.City.GeoNameID == 0 {
			return tea.NewNoContent("not found")
		}

		city = c
		return tx.Insert("", ip.String(), city, db.TTL(PositiveTTL))
	})
}

// Close the reader
func (c *Client) Close() error {
	return c.reader.Close()
}

// NewClient creates a new maxmind client
func NewClient(ctx context.Context, uri string) (*Client, error) {
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
	reader, err := geoip2.FromBytes(b)
	if err != nil {
		return nil, tea.Error(err)
	}

	c := Client{
		IPCount: int(reader.Metadata().NodeCount),
		reader:  reader,
		mapper:  ark.New(),
	}

	return &c, nil
}
