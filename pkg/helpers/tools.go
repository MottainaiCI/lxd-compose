/*

Copyright (C) 2020-2025  Daniele Rondina <geaaru@macaronios.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/MottainaiCI/lxd-compose/pkg/logger"
)

func RegexEntry(regexString string, listEntries []string) []string {
	ans := []string{}

	r := regexp.MustCompile(regexString)
	for _, e := range listEntries {

		if r != nil && r.MatchString(e) {
			ans = append(ans, e)
		}
	}
	return ans
}

func Ask(msg string) bool {
	var input string

	log := logger.GetDefaultLogger()

	log.Msg("info", false, false, msg)
	_, err := fmt.Scanln(&input)
	if err != nil {
		return false
	}
	input = strings.ToLower(input)

	if input == "y" || input == "yes" {
		return true
	}

	return false
}
