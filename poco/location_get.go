package poco

import (
	"github.com/hashicorp/go-memdb"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// Get a location by country and postalCode
func (s *LocationService) Get(id LocationId) (*Location, error) {
	locations, err := s.GetAll(id)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	return locations[0], nil
}

// GetAll fetches all locations
func (s *LocationService) GetAll(ids ...LocationId) ([]*Location, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()

	var locations []*Location
	for _, id := range ids {
		var it memdb.ResultIterator
		var err error
		switch {
		case !s.filter.Test(id.Bytes()):
			err = errors.NewNoContent()
		case id.IsCity():
			it, err = txn.Get("locations", "city", id.country, id.state, id.city)
		case id.IsCounty():
			it, err = txn.Get("locations", "county", id.country, id.state, id.county)
		case id.IsPostal():
			it, err = txn.Get("locations", "id", id.country, id.postalCode)
		case id.IsState():
			it, err = txn.Get("locations", "state", id.country, id.state)
		case id.IsCountry():
			it, err = txn.Get("locations", "country", id.country)
		}

		if err != nil {
			return nil, errors.Wrap(err)
		}

		for obj := it.Next(); obj != nil; obj = it.Next() {
			loc := obj.(Location)
			locations = append(locations, &loc)
		}
	}

	if len(locations) == 0 {
		return nil, errors.NewNoContent()
	}

	return locations, nil
}
