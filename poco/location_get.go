package poco

import (
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// Get a location by country and postalCode
func (s *LocationService) Get(country, postalCode string) (*Location, error) {
	id := NewLocationId(country, postalCode)
	if !s.cache.Test(id.Bytes()) {
		return nil, errors.NewNoContent()
	}

	raw, err := s.db.Txn(false).First("locations", "id", id.Country(), id.PostalCode())
	if err == nil && raw != nil {
		return raw.(*Location), nil
	}

	return nil, errors.NewNoContent(err)
}
