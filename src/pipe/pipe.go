package pipe

import (
	"fmt"
	"strconv"

	"github.com/nexgou/server/src/common"
)

// ParseIntPipe validates and parses a string value as an integer.
type ParseIntPipe struct{}

func (p *ParseIntPipe) Transform(value string) (any, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, common.NewBadRequestException(
			fmt.Sprintf("'%s' is not a valid integer", value),
		)
	}
	return n, nil
}

// ParseUUIDPipe validates that a string value conforms to a UUID format (36 chars).
type ParseUUIDPipe struct{}

func (p *ParseUUIDPipe) Transform(value string) (any, error) {
	if len(value) != 36 {
		return nil, common.NewBadRequestException(
			fmt.Sprintf("'%s' is not a valid UUID", value),
		)
	}
	return value, nil
}

// DefaultValuePipe returns a fallback value when the input is empty.
type DefaultValuePipe struct {
	Default string
}

func (p *DefaultValuePipe) Transform(value string) (any, error) {
	if value == "" {
		return p.Default, nil
	}
	return value, nil
}
