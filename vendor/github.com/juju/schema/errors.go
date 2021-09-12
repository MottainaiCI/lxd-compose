// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package schema

import (
	"fmt"
)

type error_ struct {
	want string
	got  interface{}
	path []string
}

func (e error_) Error() string {
	path := pathAsPrefix(e.path)
	if e.want == "" {
		return fmt.Sprintf("%sunexpected value %#v", path, e.got)
	}
	if e.got == nil {
		return fmt.Sprintf("%sexpected %s, got nothing", path, e.want)
	}
	return fmt.Sprintf("%sexpected %s, got %T(%#v)", path, e.want, e.got, e.got)
}

func parseError(path []string, expected string, err error) error {
	prefix := pathAsPrefix(path)
	return fmt.Errorf("%sconversion to %s: %s", prefix, expected, err.Error())
}
