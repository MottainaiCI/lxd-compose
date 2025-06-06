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
package specs

import (
	"gopkg.in/yaml.v3"
)

func (n *LxdCNetwork) GetName() string        { return n.Name }
func (n *LxdCNetwork) GetType() string        { return n.Type }
func (n *LxdCNetwork) GetDescription() string { return n.Description }

func NetworkFromYaml(data []byte) (*LxdCNetwork, error) {
	ans := &LxdCNetwork{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}

func (n *LxdCNetwork) IsPresentForwardAddress(a string) bool {
	ans := false
	if len(n.Forwards) > 0 {
		for idx := range n.Forwards {
			if n.Forwards[idx].ListenAddress == a {
				ans = true
				break
			}
		}
	}
	return ans
}

func (n *LxdCNetwork) GetForwardAddress(a string) *LxdCNetworkForward {
	var ans *LxdCNetworkForward = nil

	if len(n.Forwards) > 0 {
		for idx := range n.Forwards {
			if n.Forwards[idx].ListenAddress == a {
				return &n.Forwards[idx]
			}
		}
	}

	return ans
}
