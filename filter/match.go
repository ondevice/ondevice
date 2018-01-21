package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/util"
)

// operator examples
// "--with="

// list of supported operators and their implementations
var matchOperators = map[string]func(string, string) bool{
	"==": operatorEQ,
	"=":  operatorEQ,
	"!=": operatorNE,
	"<=": operatorLE,
	"<":  operatorLT,
	"<<": operatorLT,
	">=": operatorGE,
	">":  operatorGT,
	">>": operatorGT,
}

// results in 3 groups:
// - property name
// - operator (optional)
// - value (optional)
var re = regexp.MustCompile("^([a-zA-Z0-9_.\\-:]+)(?:([=!<>]{1,2})(.*))?$")

// Matches -- Returns true if the given expression is true for the device (and its properties)
//
// expr has one of the following formats:
// - <propertyName>
// - <propertyName><operator><value>
func Matches(dev api.Device, expr string) (bool, error) {
	var groups = re.FindStringSubmatch(expr)
	var key string
	var op func(string, string) bool
	var expectedValue string

	if len(groups) == 0 {
		return false, fmt.Errorf("Malformed expression: '%s'", expr)
	}
	if groups[2] != "" {
		var ok bool
		if op, ok = matchOperators[groups[2]]; !ok {
			return false, fmt.Errorf("Unsupported match operator: '%s'", groups[2])
		}
		key = groups[1]
		expectedValue = groups[3]
	} else {
		key = groups[1]
	}

	var value, ok = dev.Props[key]

	// handle special properties ('!ok' allows the server to override them explicitly)
	if !ok && strings.HasPrefix(key, "on:") {
		ok = true
		switch key {
		case "on:id":
			value = dev.ID
		case "on:name":
			value = dev.Name
		case "on:state":
			value = dev.State
		case "on:ip":
			value = dev.IP
		case "on:version":
			value = dev.Version
		case "on:createdAt":
			// uses UTC ISO 8601 dates (like "2018-01-15T17:01:33Z")
			value = util.MsecToTs(dev.CreatedAt).UTC().Format(time.RFC3339)
		default:
			ok = false
			// we can't raise an error here because otherwise server-defined special
			// properties would have to be set on ALL devices (the first one lacking it would cause this error)
			//return false, fmt.Errorf("Unknown special property: '%s'", key)
		}
	}

	if op == nil {
		// "exists"-query
		return ok && value != nil && value != "", nil
	}

	// For now we'll simply treat nil/nonexisting as the empty string, e.g.:
	// - nil == ""
	// - nil < "hello"
	// - nil != "world"
	// TODO think about nil values
	if value == nil {
		value = ""
	}

	// convert non-string values to string (to make sure we won't choke on them)
	var strVal string
	if strVal, ok = value.(string); ok {
	} else if val, ok := value.(bool); ok {
		strVal = strconv.FormatBool(val)
	} else if val, ok := value.(float64); ok {
		strVal = strconv.FormatFloat(val, 'f', -1, 64)
	} else {
		// TODO add support for extra types
		return false, fmt.Errorf("Unsupported property type: key=%s, type=%s)", key, reflect.TypeOf(value))
	}

	return op(strVal, expectedValue), nil
}

// MustMatch -- Wrapper around Matches() panicking on error
func MustMatch(dev api.Device, expr string) bool {
	var rc, err = Matches(dev, expr)
	if err != nil {
		logg.Fatalf("Error matching device properties (expr: '%s'): %s", expr, err)
	}
	return rc
}

func operatorEQ(a string, b string) bool {
	return a == b
}

func operatorGE(a string, b string) bool {
	return a >= b
}

func operatorGT(a string, b string) bool {
	return a > b
}

func operatorLE(a string, b string) bool {
	return a <= b
}

func operatorLT(a string, b string) bool {
	return a < b
}

func operatorNE(a string, b string) bool {
	return a != b
}
