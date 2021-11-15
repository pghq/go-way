package poco

import (
	"strings"
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

func (id LocationId) IsCountry() bool {
	return id.country != "" && id == Country(id.country)
}

func (id LocationId) IsState() bool {
	return id.state != "" && id == State(id.country, id.state)
}

func (id LocationId) IsCounty() bool {
	return id.county != "" && id == County(id.country, id.state, id.county)
}

func (id LocationId) IsCity() bool {
	return id.city != "" && id == City(id.country, id.state, id.city)
}

func (id LocationId) IsPostal() bool {
	return id.postalCode != "" && id == PostalCode(id.country, id.postalCode)
}

func (id LocationId) String() string {
	return strings.Join([]string{id.country, id.state, id.county, id.city, id.postalCode}, ":")
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

// LocationService is a service for retrieving locations by postal codes.
type LocationService service
