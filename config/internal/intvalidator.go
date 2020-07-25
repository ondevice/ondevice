package internal

import (
	"fmt"
	"strconv"
)

// IntValidator -- validates integer config values
type IntValidator struct {
	min, max       int
	hasMin, hasMax bool
}

// WithMax -- returns a copy of this with max limited to the given value
func (v IntValidator) WithMax(max int) IntValidator {
	v.max = max
	v.hasMax = true
	return v
}

// WithMin -- returns a copy of this with min limited to the given value
func (v IntValidator) WithMin(min int) IntValidator {
	v.min = min
	v.hasMin = true
	return v
}

// Validate -- returns nil if val is valid
func (v IntValidator) Validate(val string) error {
	return v.Value(val).Error
}

// Value -- returns a Value object for the given string
func (v IntValidator) Value(raw string) (rc Value) {
	var i int64
	if i, rc.Error = strconv.ParseInt(raw, 0, 32); rc.Error != nil {
		return
	}

	if v.hasMin && int(i) < v.min {
		rc.Error = fmt.Errorf("value is below the minimum of %d: %d", v.min, i)
	} else if v.hasMax && int(i) > v.max {
		rc.Error = fmt.Errorf("value is above the maximum of %d: %d", v.max, i)
	}

	return
}
