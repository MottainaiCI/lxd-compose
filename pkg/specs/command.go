/*

Copyright (C) 2020-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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

	"gopkg.in/yaml.v2"
)

func (c *LxdCCommand) GetName() string            { return c.Name }
func (c *LxdCCommand) GetDescription() string     { return c.Description }
func (c *LxdCCommand) GetProject() string         { return c.Project }
func (c *LxdCCommand) GetEnvs() LxdCEnvVars       { return c.Envs }
func (c *LxdCCommand) GetNodePrefix() string      { return c.NodesPrefix }
func (c *LxdCCommand) GetEnableFlags() []string   { return c.EnableFlags }
func (c *LxdCCommand) GetDisableFlags() []string  { return c.DisableFlags }
func (c *LxdCCommand) GetEnableGroups() []string  { return c.EnableGroups }
func (c *LxdCCommand) GetDisableGroups() []string { return c.DisableFlags }
func (c *LxdCCommand) GetVarFiles() []string      { return c.VarFiles }

func (c *LxdCCommand) PrepareProject(proj *LxdCProject) error {

	if proj == nil {
		return errors.New("Invalid project for apply command")
	}

	if len(c.VarFiles) > 0 {
		for _, varFile := range c.VarFiles {
			err := proj.LoadEnvVarsFile(varFile)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on load additional envs var file %s: %s",
						varFile, err.Error()),
				)
			}
		}
	}

	if len(c.Envs.EnvVars) > 0 {
		proj.AddEnvironment(&c.Envs)
	}

	return nil
}

func CommandFromYaml(data []byte) (*LxdCCommand, error) {
	ans := &LxdCCommand{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	return ans, nil
}
