package way

import (
	"net"

	"github.com/pghq/go-tea"

	"github.com/pghq/go-way/gndb"
)

// IP lookup
func (r *Radar) IP(addr string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, tea.NewError("invalid ip")
	}

	city, err := r.mdb.Get(ip)
	if err != nil {
		return nil, tea.Error(err)
	}

	loc := gndb.Location{
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

// PrimarySubdivision lookup
func (r *Radar) PrimarySubdivision(country, subdivision1 string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.gdb.Get(gndb.Primary(country, subdivision1))
}

// City lookup
func (r *Radar) City(country, subdivision1, city string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.gdb.Get(gndb.City(country, subdivision1, city))
}

// Postal lookup
func (r *Radar) Postal(country, postal string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.gdb.Get(gndb.PostalCode(country, postal))
}

// SecondarySubdivision lookup
func (r *Radar) SecondarySubdivision(country, subdivision1, subdivision2 string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.gdb.Get(gndb.Secondary(country, subdivision1, subdivision2))
}

// Country lookup
func (r *Radar) Country(country string) (*gndb.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.gdb.Get(gndb.Country(country))
}
