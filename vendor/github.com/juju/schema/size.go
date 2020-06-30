// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"reflect"
	"unicode"
)

// Size returns a Checker that accepts a string value, and returns
// the parsed string as a size in mebibytes see: https://godoc.org/github.com/juju/utils#ParseSize
func Size() Checker {
	return sizeC{}
}

type sizeC struct{}

// Coerce implements Checker Coerce method.
func (c sizeC) Coerce(v interface{}, path []string) (interface{}, error) {
	if v == nil {
		return nil, error_{"string", v, path}
	}

	typeOf := reflect.TypeOf(v).Kind()
	if typeOf != reflect.String {
		return nil, error_{"string", v, path}
	}

	value := reflect.ValueOf(v).String()
	if value == "" {
		return nil, error_{"empty string", v, path}
	}

	v, err := parseSize(value)

	if err != nil {
		return nil, err
	}

	return v, nil
}

// parseSize parses the string as a size, in mebibytes.
//
// The string must be a is a non-negative number with
// an optional multiplier suffix (M, G, T, P, E, Z, or Y).
// If the suffix is not specified, "M" is implied.
//
// Note: this function has been copied from github.com/juju/utils
// to avoid that heavy dependency.
func parseSize(str string) (MB uint64, err error) {
	// Find the first non-digit/period:
	i := strings.IndexFunc(str, func(r rune) bool {
		return r != '.' && !unicode.IsDigit(r)
	})
	var multiplier float64 = 1
	if i > 0 {
		suffix := str[i:]
		multiplier = 0
		for j := 0; j < len(sizeSuffixes); j++ {
			base := string(sizeSuffixes[j])
			// M, MB, or MiB are all valid.
			switch suffix {
			case base, base + "B", base + "iB":
				multiplier = float64(sizeSuffixMultiplier(j))
				break
			}
		}
		if multiplier == 0 {
			return 0, fmt.Errorf("invalid multiplier suffix %q, expected one of %s", suffix, []byte(sizeSuffixes))
		}
		str = str[:i]
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil || val < 0 {
		return 0, fmt.Errorf("expected a non-negative number, got %q", str)
	}
	val *= multiplier
	return uint64(math.Ceil(val)), nil
}

var sizeSuffixes = "MGTPEZY"

func sizeSuffixMultiplier(i int) int {
	return 1 << uint(i*10)
}
