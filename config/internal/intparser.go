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
func (v IntParser) WithMax(max int) IntParser {
	v.max = max
	v.hasMax = true
	return v
}

// WithMin -- returns a copy of this with min limited to the given value
func (v IntParser) WithMin(min int) IntParser {
	v.min = min
	v.hasMin = true
	return v
}

// Value -- returns a Value object for the given string
func (v IntParser) Value(raw string) (rc ValueImpl) {
	var i int64
	if i, rc.err = strconv.ParseInt(raw, 0, 32); rc.err != nil {
		return
	}

	if v.hasMin && int(i) < v.min {
		rc.err = fmt.Errorf("value is below the minimum of %d: %d", v.min, i)
	} else if v.hasMax && int(i) > v.max {
		rc.err = fmt.Errorf("value is above the maximum of %d: %d", v.max, i)
	}

	return
}
