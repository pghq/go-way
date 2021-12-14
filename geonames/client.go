package geonames

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pghq/go-ark"
	"github.com/pghq/go-ark/db"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// numColumns is expected number of columns in GeoNames csv
	numColumns = 12
)

// schema for the database
var schema = db.Schema{
	"locations": map[string][]string{
		"postal":       {"country", "postal_code"},
		"country":      {"country"},
		"subdivision1": {"country", "subdivision1"},
		"subdivision2": {"country", "subdivision1", "subdivision2"},
		"city":         {"country", "subdivision1", "city"},
	},
}

// Client for GeoNames
type Client struct {
	LocationCount int
	mapper        *ark.Mapper
}

// Get a location
func (c *Client) Get(id LocationId) (*Location, error) {
	if c == nil {
		return nil, tea.NewNoContent("client not ready")
	}

	var fence *Location
	return fence, c.mapper.Do(context.Background(), func(tx db.Txn) error {
		var query db.QueryOption
		switch {
		case id.IsCity():
			query = db.Eq("city", id.country, id.primary, id.city)
		case id.IsPostal():
			query = db.Eq("postal", id.country, id.postalCode)
		case id.IsPrimary():
			query = db.Eq("subdivision1", id.country, id.primary)
		case id.IsSecondary():
			query = db.Eq("subdivision2", id.country, id.primary, id.secondary)
		case id.IsCountry():
			query = db.Eq("country", id.country)
		default:
			return tea.NewError("bad id")
		}

		var locations []Location
		if err := tx.List("locations", &locations, query, db.Limit(-1)); err != nil {
			return tea.Error(err)
		}

		var location *Location
		for _, loc := range locations {
			l := loc
			if location == nil {
				location = &l
			} else {
				location.Add(&l)
			}
		}

		fence = location
		return nil
	}, db.ReadOnly())
}

// NewClient Creates a new GeoNames client
func NewClient(ctx context.Context, uri string) (*Client, error) {
	resp, err := client.Get(ctx, uri)
	if err != nil {
		return nil, tea.Error(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	reader := bytes.NewReader(b)
	zr, err := zip.NewReader(reader, int64(len(b)))
	if err != nil {
		return nil, tea.Error(err)
	}

	if len(zr.File) != 1 {
		return nil, tea.NewErrorf("unexpected number of files in zip, %d found", len(zr.File))
	}

	c := Client{}
	c.mapper = ark.NewRDB(schema)
	err = c.mapper.Do(ctx, func(tx db.Txn) error {
		var f io.ReadCloser
		if f, err = zr.File[0].Open(); err == nil {
			defer f.Close()
			r := csv.NewReader(f)
			r.Comma = '\t'

			var record []string
			for {
				record, err = r.Read()
				if err != nil {
					break
				}

				if len(record) != numColumns {
					return tea.NewErrorf("unexpected number of columns in csv, %d found", len(record))
				}

				latitude, err := strconv.ParseFloat(record[9], 64)
				if err != nil {
					return tea.Error(err)
				}

				longitude, err := strconv.ParseFloat(record[10], 64)
				if err != nil {
					return tea.Error(err)
				}

				country := strings.ToLower(record[0])
				postalCode := strings.ToLower(record[1])
				key := fmt.Sprintf("%s.%s", country, postalCode)
				location := Location{
					Country:      country,
					PostalCode:   postalCode,
					City:         strings.ToLower(record[2]),
					Subdivision1: strings.ToLower(record[4]),
					Subdivision2: strings.ToLower(record[5]),
					Coordinate: Coordinate{
						Latitude:  latitude,
						Longitude: longitude,
					},
				}

				if err = tx.Insert("locations", key, location); err == nil {
					c.LocationCount += 1
				}
			}
		}

		if err != nil && err != io.EOF {
			return tea.Error(err)
		}

		return nil
	}, db.BatchWrite())

	return &c, err
}
