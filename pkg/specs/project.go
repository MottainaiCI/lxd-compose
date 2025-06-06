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
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/icza/dyno"
)

func (p *LxdCProject) Init() {
	if p.Hooks == nil {
		p.Hooks = []LxdCHook{}
	}

	for idx := range p.Groups {
		p.Groups[idx].Init()
	}
}

func (p *LxdCProject) GetGroups() *[]LxdCGroup       { return &p.Groups }
func (p *LxdCProject) GetDescription() string        { return p.Description }
func (p *LxdCProject) GetName() string               { return p.Name }
func (p *LxdCProject) GetShellEnvsFilter() *[]string { return &p.ShellEnvsFilter }

func (p *LxdCProject) AddGroup(grp *LxdCGroup) {
	p.Groups = append(p.Groups, *grp)
}

func (p *LxdCProject) AddEnvironment(e *LxdCEnvVars) {
	p.Environments = append(p.Environments, *e)
}

func (p *LxdCProject) GetGroupByName(name string) *LxdCGroup {
	for idx := range p.Groups {
		if p.Groups[idx].Name == name {
			return &p.Groups[idx]
		}
	}
	return nil
}

func (p *LxdCProject) GetEnvsMap() (map[string]string, error) {
	ans := map[string]string{}

	y, err := yaml.Marshal(p.Sanitize())
	if err != nil {
		return ans, fmt.Errorf("Error on convert project %s to yaml: %s",
			p.GetName(), err.Error())
	}
	pData, err := yaml.YAMLToJSON(y)
	if err != nil {
		return ans, fmt.Errorf("Error on convert project %s to json: %s",
			p.GetName(), err.Error())
	}
	ans["project"] = string(pData)

	mfilter := make(map[string]bool, 0)
	if len(p.ShellEnvsFilter) > 0 {
		for _, k := range p.ShellEnvsFilter {
			mfilter[k] = true
		}
	}

	for _, e := range p.Environments {
		for k, v := range e.EnvVars {

			_, filtered := mfilter[k]
			if filtered {
				continue
			}

			// Bash doesn't support variable with dash.
			// I will convert dash with underscore.
			if strings.Contains(k, "-") {
				k = strings.ReplaceAll(k, "-", "_")
			}

			switch v.(type) {
			case int:
				ans[k] = fmt.Sprintf("%d", v.(int))
			case string:
				ans[k] = v.(string)
			default:
				m := dyno.ConvertMapI2MapS(v)
				y, err := yaml.Marshal(m)
				if err != nil {
					return ans, fmt.Errorf("Error on convert var %s to yaml: %s",
						k, err.Error())
				}

				data, err := yaml.YAMLToJSON(y)
				if err != nil {
					return ans, fmt.Errorf("Error on convert var %s to json: %s",
						k, err.Error())
				}
				ans[k] = string(data)
			}
		}
	}

	return ans, nil
}

func (p *LxdCProject) GetHooks(event string) []LxdCHook {
	return getHooks(&p.Hooks, event)
}

func (p *LxdCProject) GetHooks4Nodes(event string, nodes []string) []LxdCHook {
	return getHooks4Nodes(&p.Hooks, event, nodes)
}

func (p *LxdCProject) Sanitize() *LxdCProjectSanitized {
	return &LxdCProjectSanitized{
		Name:              p.Name,
		Description:       p.Description,
		IncludeGroupFiles: p.IncludeGroupFiles,
		IncludeEnvFiles:   p.IncludeEnvFiles,
		NodesPrefix:       p.NodesPrefix,
		Groups:            p.Groups,
		Hooks:             p.Hooks,
		ConfigTemplates:   p.ConfigTemplates,
	}
}

func (p *LxdCProjectSanitized) GetName() string         { return p.Name }
func (p *LxdCProjectSanitized) GetDescription() string  { return p.Description }
func (p *LxdCProjectSanitized) GetGroups() *[]LxdCGroup { return &p.Groups }

func (p *LxdCProject) GetNodesPrefix() string { return p.NodesPrefix }

func (p *LxdCProject) SetNodesPrefix(prefix string) {
	p.NodesPrefix = prefix
	for idx := range p.Groups {
		p.Groups[idx].SetNodesPrefix(prefix)
	}
}

func (p *LxdCProject) LoadEnvVarsFile(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	evars, err := EnvVarsFromYaml(content)
	if err != nil {
		return err
	}

	p.AddEnvironment(evars)

	return nil
}

func (p *LxdCProject) AddHooks(h *LxdCHooks) {
	if len(h.Hooks) > 0 {
		p.Hooks = append(p.Hooks, h.Hooks...)
	}
}
