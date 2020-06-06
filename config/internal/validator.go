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

// WithMax -- returns a copy of this with min limited to the given value
func (v IntValidator) WithMin(min int) IntValidator {
	v.min = min
	v.hasMin = true
	return v
}

// Validate -- returns nil if val is valid
func (v IntValidator) Validate(val string) error {
	var i, err = strconv.ParseInt(val, 0, 32)
	if err != nil {
		return err
	}

	if v.hasMin && int(i) < v.min {
		return fmt.Errorf("value is below the minimum of %d: %d", v.min, i)
	}

	if v.hasMax && int(i) > v.max {
		return fmt.Errorf("value is above the maximum of %d: %d", v.max, i)
	}

	return nil
}
