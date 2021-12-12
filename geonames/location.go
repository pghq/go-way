package geonames

import (
	"fmt"
	"math"
	"strings"

	"github.com/golang/geo/s2"
	"github.com/pghq/go-tea"
)

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

// Location is an instance of a GeoNames location
type Location struct {
	bounder *s2.RectBounder `db:"-"`
	Coordinate
	Country      string `db:"country"`
	PostalCode   string `db:"postal_code"`
	City         string `db:"city"`
	Subdivision1 string `db:"subdivision1"`
	Subdivision2 string `db:"subdivision2"`
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
	country    string
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

func (id LocationId) String() string {
	switch {
	case id.IsCity():
		return fmt.Sprintf("city:%s,%s,%s", id.country, id.primary, id.city)
	case id.IsSecondary():
		return fmt.Sprintf("subdivision:%s,%s,%s", id.country, id.primary, id.secondary)
	case id.IsPostal():
		return fmt.Sprintf("postal:%s,%s", id.country, id.postalCode)
	case id.IsPrimary():
		return fmt.Sprintf("subdivision:%s,%s", id.country, id.primary)
	case id.IsCountry():
		return fmt.Sprintf("city:%s", id.country)
	}

	return ""
}

// Country creates a country location id
func Country(country string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
	}
}

// Primary creates a first order subdivision location id
func Primary(country, subdivision1 string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
		primary: strings.ToLower(subdivision1),
	}
}

// Secondary creates a Secondary location id
func Secondary(country, subdivision1, subdivision2 string) LocationId {
	return LocationId{
		country:   strings.ToLower(country),
		primary:   strings.ToLower(subdivision1),
		secondary: strings.ToLower(subdivision2),
	}
}

// City creates a city location id
func City(country, primary, city string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
		primary: strings.ToLower(primary),
		city:    strings.ToLower(city),
	}
}

// PostalCode creates an instance of the postal code location id
func PostalCode(country, postalCode string) LocationId {
	return LocationId{
		country:    strings.ToLower(country),
		postalCode: postalCode,
	}
}

// ParseId parses a location string
func ParseId(s string) (LocationId, error) {
	words := strings.Split(strings.ToLower(s), ":")
	if len(words) == 2 {
		areas := strings.Split(words[1], ",")
		switch {
		case len(areas) == 1 && words[0] == "country":
			return Country(areas[0]), nil
		case len(areas) == 2 && words[0] == "subdivision":
			return Primary(areas[0], areas[1]), nil
		case len(areas) == 2 && words[0] == "postal":
			return PostalCode(areas[0], areas[1]), nil
		case len(areas) == 3 && words[0] == "subdivision":
			return Secondary(areas[0], areas[1], areas[2]), nil
		case len(areas) == 3 && words[0] == "city":
			return City(areas[0], areas[1], areas[2]), nil
		}
	}

	return LocationId{}, tea.NewErrorf("bad location %s", s)
}
