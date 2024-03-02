// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"reflect"
	"time"
)

// TimeDuration returns a Checker that accepts a string or time.Duration value,
// and returns the parsed time.Duration value.
// Empty strings are considered empty time.Duration.
func TimeDuration() Checker {
	return timeDurationC{}
}

type timeDurationC struct{}

// Coerce implements Checker Coerce method.
func (c timeDurationC) Coerce(v interface{}, path []string) (interface{}, error) {
	return asTimeDuration(v, path)
}

// TimeDurationString returns a Checker that accepts a string or time.Duration
// value, and returns the time.Duration encoded as a string. The encoding
// uses the time.Duration.String() method.
// Empty strings are considered empty time.Duration.
func TimeDurationString() Checker {
	return timeDurationStringC{}
}

type timeDurationStringC struct{}

// Coerce implements Checker Coerce method.
func (c timeDurationStringC) Coerce(v interface{}, path []string) (interface{}, error) {
	dur, err := asTimeDuration(v, path)
	if err != nil || dur == nil {
		return "", err
	}
	d, ok := dur.(time.Duration)
	if !ok {
		return "", nil
	}
	return d.String(), nil
}

func asTimeDuration(v interface{}, path []string) (interface{}, error) {
	if v == nil {
		return nil, error_{want: "string or time.Duration", got: v, path: path}
	}

	var empty time.Duration
	switch reflect.TypeOf(v).Kind() {
	case reflect.TypeOf(empty).Kind():
		return v, nil
	case reflect.String:
		vstr := reflect.ValueOf(v).String()
		if vstr == "" {
			return empty, nil
		}
		v, err := time.ParseDuration(vstr)
		if err != nil {
			return nil, parseError(path, "duration", err)
		}
		return v, nil
	default:
		return nil, error_{"string or time.Duration", v, path}
	}
}
