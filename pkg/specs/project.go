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
	"encoding/json"
)

func (p *LxdCProject) Init() {
	if p.Hooks == nil {
		p.Hooks = []LxdCHook{}
	}

	for idx, _ := range p.Groups {
		p.Groups[idx].Init()
	}
}

func (p *LxdCProject) AddGroup(grp *LxdCGroup) {
	p.Groups = append(p.Groups, *grp)
}

func (p *LxdCProject) AddEnvironment(e *LxdCEnvVars) {
	p.Environments = append(p.Environments, *e)
}

func (p *LxdCProject) GetName() string {
	return p.Name
}

func (p *LxdCProject) GetEnvsMap() map[string]string {
	ans := map[string]string{}

	for _, e := range p.Environments {
		for k, v := range e.EnvVars {
			switch v.(type) {
			case int:
				ans[k] = string(v.(int))
			case string:
				ans[k] = v.(string)
			default:
				data, err := json.Marshal(v)
				if err == nil {
					ans[k] = string(data)
				}
			}
		}
	}

	return ans
}

func (p *LxdCProject) GetHooks(event string) []LxdCHook {
	return getHooks(&p.Hooks, event)
}

func (p *LxdCProject) GetHooks4Nodes(event, node string) []LxdCHook {
	return getHooks4Nodes(&p.Hooks, event, node)
}
