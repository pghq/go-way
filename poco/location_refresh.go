package poco

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// DefaultRefreshTimeout is the default wait time for refreshing postal codes from GeoNames
	DefaultRefreshTimeout = 30 * time.Second

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

	var records [][]string
	f, err := zr.File[0].Open()
	if err == nil {
		defer f.Close()
		r := csv.NewReader(f)
		r.Comma = '\t'
		records, err = r.ReadAll()
	}

	if err != nil {
		return errors.Wrap(err)
	}

	tx := s.db.Txn(true)
	defer tx.Abort()
	cache := bloom.NewWithEstimates(uint(len(records)), BloomFalsePositiveRate)
	for _, record := range records {
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
			Country:            record[0],
			PostalCode:         record[1],
			PlaceName:          record[2],
			AdministrativeName: record[3],
			Latitude:           latitude,
			Longitude:          longitude,
		}

		if err = tx.Insert("locations", &location); err != nil {
			return errors.Wrap(err)
		}

		cache.Add(location.Id().Bytes())
	}

	tx.Commit()
	s.cache = cache
	_, _ = reader.Seek(0, io.SeekStart)
	s.source = reader

	return nil
}
