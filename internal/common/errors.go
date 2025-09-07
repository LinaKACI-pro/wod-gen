package common

import (
	"errors"
	"fmt"
)

var (
	ErrMissingBody  = errors.New("missing body")
	ErrDuration     = errors.New("duration_min must be between 15 and 120")
	ErrEmptyCatalog = errors.New("empty catalog")
	ErrNoMoves      = errors.New("no moves available")
)

type InvalidDataError struct {
	DataType string
	Data     string
}

func (e InvalidDataError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.DataType, e.Data)
}
