/*
Copyright (C) 2020-2022  Daniele Rondina <geaaru@funtoo.org>
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

func (n *LxdCAcl) GetName() string          { return n.Name }
func (n *LxdCAcl) GetDescription() string   { return n.Description }
func (n *LxdCAcl) GetDocumentation() string { return n.Documentation }

func (n *LxdCAcl) GetEgress() *[]LxdCAclRule {
	return &n.Egress
}

func (n *LxdCAcl) GetIngress() *[]LxdCAclRule {
	return &n.Ingress
}

func AclFromYaml(data []byte) (*LxdCAcl, error) {
	ans := &LxdCAcl{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}
