package poco

import (
	"bytes"
)

func (s *LocationService) Export() (*bytes.Reader, error) {
	return s.source, nil
}
