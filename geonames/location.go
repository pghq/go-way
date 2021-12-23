package geonames

import (
	"math"
	"strings"

	"github.com/golang/geo/s2"

	"github.com/pghq/go-way/country"
)

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

// Location is an instance of a GeoNames location
type Location struct {
	bounder *s2.RectBounder `db:"-"`
	Coordinate
	Country      country.Country `db:"country"`
	PostalCode   string          `db:"postal_code"`
	City         string          `db:"city"`
	Subdivision1 string          `db:"subdivision1"`
	Subdivision2 string          `db:"subdivision2"`
}

// Add to the envelope
func (l *Location) Add(loc *Location) {
	if l.bounder == nil {
		l.bounder = s2.NewRectBounder()
	}

	l.bounder.AddPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(loc.Latitude, loc.Longitude)))
}

// Center of the envelope
func (l *Location) Center() Coordinate {
	if l.bounder == nil {
		return l.Coordinate
	}

	bounder := *l.bounder
	bounder.AddPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(l.Latitude, l.Longitude)))
	center := bounder.RectBound().Center()
	return Coordinate{
		Latitude:  center.Lat.Degrees(),
		Longitude: center.Lng.Degrees(),
	}
}

// Radius of the envelope
func (l *Location) Radius() float64 {
	if l.bounder == nil {
		return 0
	}

	return math.Sqrt(l.bounder.RectBound().CapBound().Area()/math.Pi) * earthRadiusKm
}

// Coordinate represents a lat long coordinate
type Coordinate struct {
	Latitude  float64
	Longitude float64
}

// LocationId is the id for the location
type LocationId struct {
	country    country.Country
	city       string
	postalCode string
	primary    string
	secondary  string
}

// IsCountry checks if id is of country type
func (id LocationId) IsCountry() bool {
	return id.country != "" && id == Country(id.country)
}

// IsPrimary checks if id is of Primary type
func (id LocationId) IsPrimary() bool {
	return id.primary != "" && id == Primary(id.country, id.primary)
}

// IsSecondary checks if id is of Secondary type
func (id LocationId) IsSecondary() bool {
	return id.secondary != "" && id == Secondary(id.country, id.primary, id.secondary)
}

// IsCity checks if id is of city type
func (id LocationId) IsCity() bool {
	return id.city != "" && id == City(id.country, id.primary, id.city)
}

// IsPostal checks if id is of postal type
func (id LocationId) IsPostal() bool {
	return id.postalCode != "" && id == PostalCode(id.country, id.postalCode)
}

// Country creates a country location id
func Country(country country.Country) LocationId {
	return LocationId{
		country: country,
	}
}

// Primary creates a first order subdivision location id
func Primary(country country.Country, subdivision1 string) LocationId {
	return LocationId{
		country: country,
		primary: strings.ToLower(subdivision1),
	}
}

// Secondary creates a Secondary location id
func Secondary(country country.Country, subdivision1, subdivision2 string) LocationId {
	return LocationId{
		country:   country,
		primary:   strings.ToLower(subdivision1),
		secondary: strings.ToLower(subdivision2),
	}
}

// City creates a city location id
func City(country country.Country, primary, city string) LocationId {
	return LocationId{
		country: country,
		primary: strings.ToLower(primary),
		city:    strings.ToLower(city),
	}
}

// PostalCode creates an instance of the postal code location id
func PostalCode(country country.Country, postalCode string) LocationId {
	return LocationId{
		country:    country,
		postalCode: postalCode,
	}
}
