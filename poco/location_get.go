package poco

import (
	"github.com/hashicorp/go-memdb"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// Get a location by country and postalCode
func (s *LocationService) Get(key LocationKey) (*Location, error) {
	if !s.cache.Test(key.Bytes()) {
		return nil, errors.NewNoContent()
	}

	raw, err := s.db.Txn(false).First("locations", "id", key.Country, key.PostalCode)
	if err != nil {
		if errors.Is(err, memdb.ErrNotFound) {
			return nil, errors.NewNoContent(err)
		}
		return nil, errors.Wrap(err)
	}

	return raw.(*Location), nil
}
