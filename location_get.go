package way

import (
	"net"

	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/geonames"
)

// IP lookup
func (s *LocationService) IP(addr string) (*geonames.Location, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, tea.NewError("invalid ip")
	}

	city, err := s.mdb.Get(ip)
	if err != nil {
		return nil, tea.Error(err)
	}

	loc := geonames.Location{
		Country:    city.Country.IsoCode,
		PostalCode: city.Postal.Code,
	}

	if city, present := city.City.Names["en"]; present {
		loc.City = city
	}

	if len(city.Subdivisions) > 0 {
		loc.Subdivision1 = city.Subdivisions[0].IsoCode
	}

	loc.Latitude = city.Location.Latitude
	loc.Longitude = city.Location.Longitude

	return &loc, nil
}

// Primary subdivision lookup
func (s *LocationService) Primary(country, subdivision1 string) (*geonames.Location, error) {
	return s.gdb.Get(geonames.Primary(country, subdivision1))
}

// City lookup
func (s *LocationService) City(country, subdivision1, city string) (*geonames.Location, error) {
	return s.gdb.Get(geonames.City(country, subdivision1, city))
}

// Postal lookup
func (s *LocationService) Postal(country, postal string) (*geonames.Location, error) {
	return s.gdb.Get(geonames.PostalCode(country, postal))
}

// Secondary subdivision lookup
func (s *LocationService) Secondary(country, subdivision1, subdivision2 string) (*geonames.Location, error) {
	return s.gdb.Get(geonames.Secondary(country, subdivision1, subdivision2))
}

// Country lookup
func (s *LocationService) Country(country string) (*geonames.Location, error) {
	return s.gdb.Get(geonames.Country(country))
}
