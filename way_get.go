package way

import (
	"net"

	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/geonames"
)

// IP lookup
func (r *Radar) IP(addr string) (*geonames.Location, error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, tea.NewError("invalid ip")
	}

	city, err := r.maxmind.Get(ip)
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

// PSD primary subdivision lookup
func (r *Radar) PSD(country, subdivision1 string) (*geonames.Location, error) {
	return r.geonames.Get(geonames.Primary(country, subdivision1))
}

// City lookup
func (r *Radar) City(country, subdivision1, city string) (*geonames.Location, error) {
	return r.geonames.Get(geonames.City(country, subdivision1, city))
}

// Postal lookup
func (r *Radar) Postal(country, postal string) (*geonames.Location, error) {
	return r.geonames.Get(geonames.PostalCode(country, postal))
}

// SSD secondary division lookup
func (r *Radar) SSD(country, subdivision1, subdivision2 string) (*geonames.Location, error) {
	return r.geonames.Get(geonames.Secondary(country, subdivision1, subdivision2))
}

// Country lookup
func (r *Radar) Country(country string) (*geonames.Location, error) {
	return r.geonames.Get(geonames.Country(country))
}
