package poco

import (
	"bytes"
)

// Export csv data of all locations since last refresh
func (s *LocationService) Export() (*bytes.Reader, error) {
	return s.source, nil
}
