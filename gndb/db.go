package gndb

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pghq/go-ark"
	"github.com/pghq/go-ark/rdb"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// numColumns is expected number of columns in GeoNames csv
	numColumns = 12
)

// schema is the schema for memdb
var schema = rdb.Schema{
	"locations": map[string][]string{
		"primary":      {"country", "postal_code"},
		"country":      {"country"},
		"subdivision1": {"country", "subdivision1"},
		"subdivision2": {"country", "subdivision1", "subdivision2"},
		"city":         {"country", "subdivision1", "city"},
	},
}

// DB for GeoNames
type DB struct {
	Size int
	conn *ark.RDBConn
}

// Get for locations
func (db *DB) Get(id LocationId) (*Location, error) {
	if db == nil {
		return nil, tea.NewNoContent("db not ready")
	}

	var fence *Location
	return fence, db.conn.Do(context.Background(), func(tx *ark.RDBTxn) error {
		f := rdb.Ft()
		switch {
		case id.IsCity():
			f = f.IdxEq("city", id.country, id.primary, id.city)
		case id.IsPostal():
			f = f.IdxEq("primary", id.country, id.postalCode)
			var location *Location
			if _, err := tx.Get(ark.Qy().Table("locations").Filter(f), &location).Resolve(); err != nil {
				return tea.Error(err)
			}
			fence = location
			return nil
		case id.IsPrimary():
			f = f.IdxEq("subdivision1", id.country, id.primary)
		case id.IsSecondary():
			f = f.IdxEq("subdivision2", id.country, id.primary, id.secondary)
		case id.IsCountry():
			f = f.IdxEq("country", id.country)
		default:
			return tea.NewError("bad id")
		}

		var locations []Location
		if _, err := tx.List(ark.Qy().Table("locations").Filter(f), &locations).Resolve(); err != nil {
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
	}, true)
}

// Open geonames db
func Open(ctx context.Context, uri string) (*DB, error) {
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

	db := DB{}
	db.conn, _ = ark.Open().DSN(schema).ConnectRDB(ctx, "inmem")
	err = db.conn.Do(ctx, func(tx *ark.RDBTxn) error {
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

				location := Location{
					Country:      strings.ToLower(record[0]),
					PostalCode:   strings.ToLower(record[1]),
					City:         strings.ToLower(record[2]),
					Subdivision1: strings.ToLower(record[4]),
					Subdivision2: strings.ToLower(record[5]),
					Coordinate: Coordinate{
						Latitude:  latitude,
						Longitude: longitude,
					},
				}

				if _, err = tx.Insert("locations", location).Resolve(); err == nil {
					db.Size += 1
				}
			}
		}

		if err != nil && err != io.EOF {
			return tea.Error(err)
		}

		return nil
	})

	return &db, err
}
