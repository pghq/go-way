package poco

import (
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

// Id is the location id
func (l *Location) Id() *LocationId {
	return NewLocationId(l.Country, l.PostalCode)
}

// LocationId is the id for the location
type LocationId struct {
	country    string
	postalCode string
}

// NewLocationId creates an instance of the location id
func NewLocationId(country, postalCode string) *LocationId{
	return &LocationId{
		country:    country,
		postalCode: postalCode,
	}
}

// Country gets the normalized country code for the location
func (id *LocationId) Country() string{
	return strings.ToLower(id.country)
}

// PostalCode gets the normalized postal code for the location
func (id *LocationId) PostalCode() string{
	return strings.ToLower(id.postalCode)
}

func (id *LocationId) String() string {
	return id.Country() + id.PostalCode()
}

func (id *LocationId) Bytes() []byte {
	return []byte(id.String())
}

// LocationService is a service for retrieving locations by postal codes.
type LocationService service
