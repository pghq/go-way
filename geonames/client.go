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
	"github.com/pghq/go-ark/database"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// numColumns is expected number of columns in GeoNames csv
	numColumns = 12
)

// schema for the database
var schema = database.Schema{
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
	db            *ark.Mapper
}

// Get a location
func (c *Client) Get(id LocationId) (*Location, error) {
	if c == nil {
		return nil, tea.ErrNotFound("client not ready")
	}

	var fence *Location
	return fence, c.db.View(context.Background(), func(tx ark.Txn) error {
		var query database.QueryOption
		switch {
		case id.IsCity():
			query = database.Eq("city", id.country, id.primary, id.city)
		case id.IsPostal():
			query = database.Eq("postal", id.country, id.postalCode)
		case id.IsPrimary():
			query = database.Eq("subdivision1", id.country, id.primary)
		case id.IsSecondary():
			query = database.Eq("subdivision2", id.country, id.primary, id.secondary)
		case id.IsCountry():
			query = database.Eq("country", id.country)
		default:
			return tea.Err("bad id")
		}

		var locations []Location
		if err := tx.List("locations", &locations, query, database.Limit(-1)); err != nil {
			return tea.Stack(err)
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
	})
}

// NewClient Creates a new GeoNames client
func NewClient(ctx context.Context, uri string) (*Client, error) {
	resp, err := client.Get(ctx, uri)
	if err != nil {
		return nil, tea.Stack(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	reader := bytes.NewReader(b)
	zr, err := zip.NewReader(reader, int64(len(b)))
	if err != nil {
		return nil, tea.Stack(err)
	}

	if len(zr.File) != 1 {
		return nil, tea.Errf("unexpected number of files in zip, %d found", len(zr.File))
	}

	c := Client{}
	c.db = ark.New("memory://", database.Storage(schema))
	err = c.db.Do(ctx, func(tx ark.Txn) error {
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
					return tea.Errf("unexpected number of columns in csv, %d found", len(record))
				}

				latitude, err := strconv.ParseFloat(record[9], 64)
				if err != nil {
					return tea.Stack(err)
				}

				longitude, err := strconv.ParseFloat(record[10], 64)
				if err != nil {
					return tea.Stack(err)
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
			return tea.Stack(err)
		}

		return nil
	}, database.BatchWrite())

	return &c, err
}
