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
	"gopkg.in/yaml.v2"
)

func (g *LxdCGroup) Init() {
	// Initialize Hooks array to reduce code checks.
	if g.Hooks == nil {
		g.Hooks = []LxdCHook{}
	}

	for idx, _ := range g.Nodes {
		g.Nodes[idx].Init()
	}
}

func GroupFromYaml(data []byte) (*LxdCGroup, error) {
	ans := &LxdCGroup{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}

func (g *LxdCGroup) GetNodesPrefix() string { return g.NodesPrefix }

func (g *LxdCGroup) SetNodesPrefix(prefix string) {
	g.NodesPrefix = prefix

	for idx, _ := range g.Nodes {
		if g.Nodes[idx].NamePrefix == "" {
			g.Nodes[idx].NamePrefix = prefix
		}
	}
}

func (g *LxdCGroup) GetHooks(event string) []LxdCHook {
	return getHooks(&g.Hooks, event)
}

func (g *LxdCGroup) GetHooks4Nodes(event string, nodes []string) []LxdCHook {
	return getHooks4Nodes(&g.Hooks, event, nodes)
}

func (g *LxdCGroup) ToProcess(groupsEnabled, groupsDisabled []string) bool {
	ans := false

	if len(groupsDisabled) > 0 {
		for _, gd := range groupsDisabled {
			if gd == g.Name {
				return false
			}
		}
	}

	if len(groupsEnabled) > 0 {
		for _, ge := range groupsEnabled {
			if ge == g.Name {
				ans = true
				break
			}
		}
	} else {
		ans = true
	}

	return ans
}
