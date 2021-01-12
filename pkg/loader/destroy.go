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
package loader

import (
	"errors"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

func (i *LxdCInstance) DestroyProject(projectName string) error {

	env := i.GetEnvByProjectName(projectName)
	if env == nil {
		return errors.New("No environment found for project " + projectName)
	}

	proj := env.GetProjectByName(projectName)
	if proj == nil {
		return errors.New("No project found with name " + projectName)
	}

	if i.NodesPrefix != "" {
		proj.SetNodesPrefix(i.NodesPrefix)
	}

	for _, grp := range proj.Groups {

		err := i.DestroyGroup(&grp, proj, env)
		if err != nil {
			return err
		}

	}

	return nil
}

func (i *LxdCInstance) DestroyGroup(group *specs.LxdCGroup, proj *specs.LxdCProject, env *specs.LxdCEnvironment) error {

	// Initialize executor
	executor := executor.NewLxdCExecutor(group.Connection,
		i.Config.GetGeneral().LxdConfDir, []string{}, group.Ephemeral,
		i.Config.GetLogging().CmdsOutput,
		i.Config.GetLogging().RuntimeCmdsOutput)
	err := executor.Setup()
	if err != nil {
		i.Logger.Error("Error on initialize executor for group " + group.Name + ": " + err.Error())
		return err
	}

	for _, node := range group.Nodes {

		isPresent, err := executor.IsPresentContainer(node.GetName())
		if err != nil {
			i.Logger.Error("Error on check if container " + node.GetName() +
				" is present: " + err.Error())
			return err
		}

		if isPresent {
			err = executor.DeleteContainer(node.GetName())
			if err != nil {
				i.Logger.Error("Error on destroy container " + node.GetName() +
					": " + err.Error())
				return err
			}
		}

	}

	return nil
}
