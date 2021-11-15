package poco

import (
	"math"

	"github.com/golang/geo/s2"
	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

// Envelope is an instance of the minimum bounding rectangle (approx.)
type Envelope struct {
	locations []*Location
	bounder   *s2.RectBounder
}

// AddLocation to the envelope
func (e *Envelope) AddLocation(loc *Location) {
	e.locations = append(e.locations, loc)
	e.bounder.AddPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(loc.Latitude, loc.Longitude)))
}

// Locations within the envelope
func (e *Envelope) Locations() []*Location {
	return e.locations
}

// Center of the envelope
func (e *Envelope) Center() Coordinate {
	center := e.bounder.RectBound().Center()
	return Coordinate{
		Latitude:  center.Lat.Degrees(),
		Longitude: center.Lng.Degrees(),
	}
}

// Radius of the envelope
func (e *Envelope) Radius() float64 {
	return math.Sqrt(e.bounder.RectBound().CapBound().Area()/math.Pi) * earthRadiusKm
}

// Envelope retrieves the minimum bounding rectangle of a set of postal codes
func (s *LocationService) Envelope(ids ...LocationId) (*Envelope, error) {
	envelope := Envelope{
		bounder: s2.NewRectBounder(),
	}

	locations, err := s.GetAll(ids...)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	for _, loc := range locations {
		envelope.AddLocation(loc)
	}

	return &envelope, nil
}
