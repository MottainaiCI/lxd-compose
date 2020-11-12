/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
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
package specs

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

func EnvVarsFromYaml(data []byte) (*LxdCEnvVars, error) {
	ans := &LxdCEnvVars{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	return ans, nil
}

func NewEnvVars() *LxdCEnvVars {
	return &LxdCEnvVars{
		EnvVars: make(map[string]interface{}, 0),
	}
}

func (e *LxdCEnvVars) AddKVAggregated(aggregatedEnv string) error {

	if aggregatedEnv == "" {
		return errors.New("Invalid key")
	}

	if strings.Index(aggregatedEnv, "=") < 0 {
		return errors.New(fmt.Sprintf("Invalid KV %s without =.", aggregatedEnv))
	}

	key := aggregatedEnv[0:strings.Index(aggregatedEnv, "=")]
	value := aggregatedEnv[strings.Index(aggregatedEnv, "=")+1:]

	e.EnvVars[key] = value

	return nil
}

func (e *LxdCEnvVars) AddKV(key, value string) error {
	if key == "" {
		return errors.New("Invalid key")
	}

	if value == "" {
		return errors.New(fmt.Sprintf("Invalid value for key %s", key))
	}

	e.EnvVars[key] = value

	return nil
}
