package poco

import (
	"fmt"
	"strings"
)

// Location is an instance of a GeoNames location
type Location struct {
	Country            string
	PostalCode         string
	PlaceName          string
	AdministrativeName string
	Latitude           float64
	Longitude          float64
}

func (l *Location) Key() *LocationKey {
	return &LocationKey{
		Country:    l.Country,
		PostalCode: l.PostalCode,
	}
}

type LocationKey struct {
	Country    string
	PostalCode string
}

func (k *LocationKey) String() string {
	return fmt.Sprintf("%s:%s", strings.ToLower(k.Country), strings.ToLower(k.PostalCode))
}

func (k *LocationKey) Bytes() []byte {
	return []byte(k.String())
}

// LocationService is a service for retrieving locations by postal codes.
type LocationService service
