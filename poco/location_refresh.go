package poco

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

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// DefaultRefreshTimeout is the default wait time for refreshing postal codes from GeoNames
	DefaultRefreshTimeout = 60 * time.Second

	// DefaultRefreshLocation is the default origin location for the GeoName export
	DefaultRefreshLocation = "https://download.geonames.org/export/zip/allCountries.zip"

	// NumColumns is expected number of columns in GeoNames csv
	NumColumns = 12

	// BloomFalsePositiveRate is the false positive rate for the bloom filter
	BloomFalsePositiveRate = 0.01
)

// Refresh postal codes from GeoNames and cache
func (s *LocationService) Refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.conf.refreshTimeout)
	defer cancel()

	resp, err := s.client.Get(ctx, s.conf.refreshLocation)
	if err != nil {
		return errors.Wrap(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	reader := bytes.NewReader(b)
	zr, err := zip.NewReader(reader, int64(len(b)))
	if err != nil {
		return errors.Wrap(err)
	}

	if len(zr.File) != 1 {
		return errors.Newf("unexpected number of files in zip, %d found", len(zr.File))
	}

	tx := s.db.Txn(true)
	defer tx.Abort()

	locations := make(map[string]*Location)
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
				return errors.Newf("unexpected number of columns in csv, %d found", len(record))
			}

			latitude, err := strconv.ParseFloat(record[9], 64)
			if err != nil {
				return errors.Wrap(err)
			}

			longitude, err := strconv.ParseFloat(record[10], 64)
			if err != nil {
				return errors.Wrap(err)
			}

			location := Location{
				Country:    strings.ToLower(record[0]),
				PostalCode: strings.ToLower(record[1]),
				City:       strings.ToLower(record[2]),
				State:      strings.ToLower(record[4]),
				County:     strings.ToLower(record[5]),
				Coordinate: Coordinate{
					Latitude:  latitude,
					Longitude: longitude,
				},
			}

			if err = tx.Insert("locations", location); err != nil {
				return errors.Wrap(err)
			}

			locations[location.Id().String()] = &location
		}
	}

	if err != nil && err != io.EOF {
		return errors.Wrap(err)
	}

	filter := bloom.NewWithEstimates(uint(len(locations)), BloomFalsePositiveRate)
	for _, location := range locations {
		filter.Add(location.Id().Bytes())
		filter.Add(location.CountryId().Bytes())
		filter.Add(location.CountyId().Bytes())
		filter.Add(location.StateId().Bytes())
		filter.Add(location.CityId().Bytes())
	}

	_, _ = reader.Seek(0, io.SeekStart)

	tx.Commit()
	s.filter = filter
	s.source = reader

	return nil
}
