package poco

import (
	"fmt"
	"strings"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// Location is an instance of a GeoNames location
type Location struct {
	Coordinate
	Country    string
	PostalCode string
	City       string
	County     string
	State      string
}

// CountryId is the country id
func (l *Location) CountryId() LocationId {
	return Country(l.Country)
}

// StateId is the state id
func (l *Location) StateId() LocationId {
	return State(l.Country, l.State)
}

// CountyId is the county id
func (l *Location) CountyId() LocationId {
	return County(l.Country, l.State, l.County)
}

// CityId is the city id
func (l *Location) CityId() LocationId {
	return City(l.Country, l.State, l.City)
}

// Id is the location id
func (l *Location) Id() LocationId {
	return PostalCode(l.Country, l.PostalCode)
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
	state      string
	postalCode string
	county     string
}

// IsCountry checks if id is of country type
func (id LocationId) IsCountry() bool {
	return id.country != "" && id == Country(id.country)
}

// IsState checks if id is of state type
func (id LocationId) IsState() bool {
	return id.state != "" && id == State(id.country, id.state)
}

// IsCounty checks if id is of county type
func (id LocationId) IsCounty() bool {
	return id.county != "" && id == County(id.country, id.state, id.county)
}

// IsCity checks if id is of city type
func (id LocationId) IsCity() bool {
	return id.city != "" && id == City(id.country, id.state, id.city)
}

// IsPostal checks if id is of postal type
func (id LocationId) IsPostal() bool {
	return id.postalCode != "" && id == PostalCode(id.country, id.postalCode)
}

func (id LocationId) String() string {
	switch {
	case id.IsCity():
		return fmt.Sprintf("city:%s,%s,%s", id.country, id.state, id.city)
	case id.IsCounty():
		return fmt.Sprintf("county:%s,%s,%s", id.country, id.state, id.county)
	case id.IsPostal():
		return fmt.Sprintf("postal:%s,%s", id.country, id.postalCode)
	case id.IsState():
		return fmt.Sprintf("city:%s,%s", id.country, id.state)
	case id.IsCountry():
		return fmt.Sprintf("city:%s", id.country)
	}

	return ""
}

func (id LocationId) Bytes() []byte {
	return []byte(id.String())
}

// Country creates a country location id
func Country(country string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
	}
}

// State creates a state location id
func State(country, state string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
		state:   strings.ToLower(state),
	}
}

// County creates a county location id
func County(country, state, county string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
		state:   strings.ToLower(state),
		county:  strings.ToLower(county),
	}
}

// City creates a city location id
func City(country, state, city string) LocationId {
	return LocationId{
		country: strings.ToLower(country),
		state:   strings.ToLower(state),
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

// ParseLocation parses a location string
func ParseLocation(s string) (LocationId, error) {
	words := strings.Split(strings.ToLower(s), ":")
	if len(words) == 2 {
		areas := strings.Split(words[1], ",")
		switch {
		case len(areas) == 1 && words[0] == "country":
			return Country(areas[0]), nil
		case len(areas) == 2 && words[0] == "state":
			return State(areas[0], areas[1]), nil
		case len(areas) == 2 && words[0] == "postal":
			return PostalCode(areas[0], areas[1]), nil
		case len(areas) == 3 && words[0] == "county":
			return County(areas[0], areas[1], areas[2]), nil
		case len(areas) == 3 && words[0] == "city":
			return City(areas[0], areas[1], areas[2]), nil
		}
	}

	return LocationId{}, errors.Newf("bad location %s", s)
}

// LocationService is a service for retrieving locations by postal codes.
type LocationService service
