package geonames

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/hashicorp/go-memdb"
	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/client"
)

const (
	// NumColumns is expected number of columns in GeoNames csv
	NumColumns = 12

	// positiveTTL is the positive ttl for search queries
	positiveTTL = 30 * time.Minute

	// negativeTTL is the negative ttl for search queries
	negativeTTL = 90 * time.Minute
)

// schema is the schema for memdb
var schema = memdb.DBSchema{
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
				"subdivision1": {
					Name:         "subdivision1",
					AllowMissing: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Country"},
							&memdb.StringFieldIndex{Field: "Subdivision1"},
						},
					},
				},
				"subdivision2": {
					Name:         "subdivision2",
					AllowMissing: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Country"},
							&memdb.StringFieldIndex{Field: "Subdivision1"},
							&memdb.StringFieldIndex{Field: "Subdivision2"},
						},
					},
				},
				"city": {
					Name:         "city",
					AllowMissing: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "Country"},
							&memdb.StringFieldIndex{Field: "Subdivision1"},
							&memdb.StringFieldIndex{Field: "City"},
						},
					},
				},
			},
		},
	},
}

// DB for geonames
type DB struct {
	mdb    *memdb.MemDB
	cache  *ristretto.Cache
	source *bytes.Reader
}

// Get for locations
func (db *DB) Get(id LocationId) (*Location, error) {
	v, present := db.cache.Get(id.String())
	if present {
		if err, ok := v.(error); ok {
			return nil, tea.Error(err)
		}

		if loc, ok := v.(*Location); ok {
			return loc, nil
		}
	}

	txn := db.mdb.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	var err error
	switch {
	case id.IsCity():
		it, err = txn.Get("locations", "city", id.country, id.Primary, id.city)
	case id.IsPostal():
		it, err = txn.Get("locations", "id", id.country, id.postalCode)
	case id.IsPrimary():
		it, err = txn.Get("locations", "subdivision1", id.country, id.Primary)
	case id.IsSecondary():
		it, err = txn.Get("locations", "subdivision2", id.country, id.Primary, id.Secondary)
	case id.IsCountry():
		it, err = txn.Get("locations", "country", id.country)
	default:
		err = tea.NewError("bad id")
	}

	if err != nil {
		return nil, tea.Error(err)
	}

	var location *Location
	for obj := it.Next(); obj != nil; obj = it.Next() {
		loc := obj.(Location)
		if location == nil {
			location = &loc
		}
		location.Add(&loc)
	}

	if location == nil {
		err := tea.NewNoContent()
		db.cache.SetWithTTL(id.String(), err, 1, negativeTTL)
		return nil, err
	}

	db.cache.SetWithTTL(id.String(), location, 1, positiveTTL)

	return location, nil
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
	// at this point we have a valid schema
	db.mdb, _ = memdb.NewMemDB(&schema)

	tx := db.mdb.Txn(true)
	defer tx.Abort()

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

			if len(record) != NumColumns {
				return nil, tea.NewErrorf("unexpected number of columns in csv, %d found", len(record))
			}

			latitude, err := strconv.ParseFloat(record[9], 64)
			if err != nil {
				return nil, tea.Error(err)
			}

			longitude, err := strconv.ParseFloat(record[10], 64)
			if err != nil {
				return nil, tea.Error(err)
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

			if err = tx.Insert("locations", location); err != nil {
				return nil, tea.Error(err)
			}
		}
	}

	if err != nil && err != io.EOF {
		return nil, tea.Error(err)
	}

	_, _ = reader.Seek(0, io.SeekStart)

	tx.Commit()
	db.cache, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	db.source = reader

	return &db, nil
}
