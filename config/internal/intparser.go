package internal

import (
	"fmt"
	"strconv"
)

// IntParser -- validates integer config values
type IntParser struct {
	min, max       int
	hasMin, hasMax bool
}

// WithMax -- returns a copy of this with max limited to the given value
func (p IntParser) WithMax(max int) IntParser {
	p.max = max
	p.hasMax = true
	return p
}

// WithMin -- returns a copy of this with min limited to the given value
func (p IntParser) WithMin(min int) IntParser {
	p.min = min
	p.hasMin = true
	return p
}

// Value -- returns a Value object for the given string
func (p IntParser) Value(raw string) ValueImpl {
	var rc = ValueImpl{parser: p}
	var i int64
	if i, rc.err = strconv.ParseInt(raw, 0, 32); rc.err != nil {
		return rc
	}

	if p.hasMin && int(i) < p.min {
		rc.err = fmt.Errorf("value is below the minimum of %d: %d", p.min, i)
	} else if p.hasMax && int(i) > p.max {
		rc.err = fmt.Errorf("value is above the maximum of %d: %d", p.max, i)
	}

	return rc
}
