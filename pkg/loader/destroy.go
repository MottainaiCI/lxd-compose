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
package loader

import (
	"errors"
	"fmt"

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

	// Retrieve pre-project-shutdown/post-project-shutdown hooks
	preProjHooks := proj.GetHooks4Nodes("pre-project-shutdown", []string{"*"})
	postProjHooks := proj.GetHooks4Nodes("post-project-shutdown", []string{"*"})

	// Execute pre-project hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s] Running %d pre-project-shutdown hooks... ", projectName, len(preProjHooks)))
	err := i.ProcessHooks(&preProjHooks, proj, nil, nil)
	if err != nil {
		return err
	}

	for _, grp := range proj.Groups {

		if !grp.ToProcess(i.GroupsEnabled, i.GroupsDisabled) {
			i.Logger.Debug("Skipped group ", grp.Name)
			continue
		}

		err := i.DestroyGroup(&grp, proj, env)
		if err != nil {
			return err
		}

	}

	// Execute pre-project hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s] Running %d post-project-shutdown hooks... ", projectName, len(postProjHooks)))
	err = i.ProcessHooks(&postProjHooks, proj, nil, nil)

	return err
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

	// Retrieve pre-group hooks from project
	preGroupHooks := proj.GetHooks4Nodes("pre-group-shutdown", []string{"*"})
	// Retrieve pre-group hooks from group
	preGroupHooks = append(preGroupHooks, group.GetHooks4Nodes("pre-group-shutdown", []string{"*"})...)

	// Run pre-group hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d pre-group-shtudown hooks... ", proj.Name, group.Name, len(preGroupHooks)))
	err = i.ProcessHooks(&preGroupHooks, proj, group, nil)
	if err != nil {
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

			// Retrieve and run pre-node-shutdown hooks of the node from project
			preNodeShutdownHooks := i.GetNodeHooks4Event("pre-node-shutdown", proj, group, &node)
			err = i.ProcessHooks(&preNodeShutdownHooks, proj, group, &node)
			if err != nil {
				return err
			}

			err = executor.DeleteContainer(node.GetName())
			if err != nil {
				i.Logger.Error("Error on destroy container " + node.GetName() +
					": " + err.Error())
				return err
			}

			// Retrieve and run post-node-shutdown hooks of the node from project
			postNodeShutdownHooks := i.GetNodeHooks4Event("post-node-shutdown", proj, group, &node)
			err = i.ProcessHooks(&postNodeShutdownHooks, proj, group, &node)
			if err != nil {
				return err
			}

		}

	}

	// Retrieve post-group hooks from project
	postGroupHooks := proj.GetHooks4Nodes("post-group-shutdown", []string{"*"})
	postGroupHooks = append(postGroupHooks, group.GetHooks4Nodes("post-group-shutdown", []string{"*"})...)

	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d post-group-shutdown hooks... ", proj.Name, group.Name, len(postGroupHooks)))
	err = i.ProcessHooks(&postGroupHooks, proj, group, nil)

	return err
}
