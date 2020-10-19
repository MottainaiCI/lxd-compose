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
	"fmt"
	"path/filepath"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
	"github.com/MottainaiCI/lxd-compose/pkg/template"
)

func (i *LxdCInstance) GetNodeHooks4Event(event string, proj *specs.LxdCProject, group *specs.LxdCGroup, node *specs.LxdCNode) []specs.LxdCHook {

	// Retrieve project hooks
	projHooks := proj.GetHooks4Nodes(event, "*")
	projHooks = specs.FilterHooks4Node(&projHooks, node.Name)

	// Retrieve group hooks
	groupHooks := group.GetHooks4Nodes(event, "*")
	groupHooks = specs.FilterHooks4Node(&groupHooks, node.Name)

	ans := projHooks
	ans = append(ans, groupHooks...)
	ans = append(ans, node.GetAllHooks(event)...)

	return ans
}

func (i *LxdCInstance) ApplyProject(projectName string) error {

	env := i.GetEnvByProjectName(projectName)
	if env == nil {
		return errors.New("No environment found for project " + projectName)
	}

	proj := env.GetProjectByName(projectName)
	if proj == nil {
		return errors.New("No project found with name " + projectName)
	}

	// Get only host hooks. All other hooks are handled by group and node.
	preProjHooks := proj.GetHooks4Nodes("pre-project", "host")
	postProjHooks := proj.GetHooks4Nodes("post-project", "*")

	// Create executor for host commands.
	executor := executor.NewLxdCExecutor("local",
		i.Config.GetGeneral().LxdConfDir, []string{}, true,
		i.Config.GetLogging().CmdsOutput)
	// Setup is not needed here.

	// Execute pre-project hooks
	err := i.ProcessHooks(&preProjHooks, executor, proj, nil)

	compiler, err := template.NewProjectTemplateCompiler(env, proj)
	if err != nil {
		return err
	}

	// Compiler project files
	err = template.CompileProjectFiles(proj, compiler, template.CompilerOpts{})
	if err != nil {
		return err
	}

	for _, grp := range proj.Groups {

		err := i.ApplyGroup(&grp, proj, env, compiler)
		if err != nil {
			return err
		}

	}

	// Execute post-project hooks
	err = i.ProcessHooks(&postProjHooks, executor, proj, nil)
	if err != nil {
		return err
	}

	return nil
}

func (i *LxdCInstance) ProcessHooks(hooks *[]specs.LxdCHook, executor *executor.LxdCExecutor, proj *specs.LxdCProject, group *specs.LxdCGroup) error {
	var res int
	nodes := []specs.LxdCNode{}
	storeVar := false

	if len(*hooks) > 0 {

		runSingleCmd := func(h *specs.LxdCHook, node, cmds string) error {

			if h.Out2Var != "" || h.Err2Var != "" {
				storeVar = true
			} else {
				storeVar = false
			}

			envs, err := proj.GetEnvsMap()
			if err != nil {
				return err
			}
			if _, ok := envs["HOME"]; !ok {
				envs["HOME"] = "/"
			}

			if h.Node == "host" {
				if storeVar {
					res, err = executor.RunHostCommandWithOutput4Var(cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint)
				} else {
					res, err = executor.RunHostCommand(cmds, envs, h.Entrypoint)
				}
			} else {

				if node != "" {
					_, _, _, nodeEntity := i.GetEntitiesByNodeName(node)
					if nodeEntity == nil {
						return errors.New("Error on retrieve node object for name " + node)
					}

					json, err := nodeEntity.ToJson()
					if err != nil {
						return err
					}
					envs["node"] = json

					if len(nodeEntity.Labels) > 0 {
						for k, v := range nodeEntity.Labels {
							envs[k] = v
						}
					}
				}

				if storeVar {
					res, err = executor.RunCommandWithOutput4Var(node, cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint)
				} else {
					res, err = executor.RunCommand(node, cmds, envs, h.Entrypoint)
				}

			}

			if err != nil {
				i.Logger.Error("Error " + err.Error())
				return err
			}

			if res != 0 {
				i.Logger.Error(fmt.Sprintf("Command result wrong (%d). Exiting.", res))
				return errors.New("Error on execute command: " + cmds)
			}

			if storeVar {
				if len(proj.Environments) == 0 {
					proj.AddEnvironment(&specs.LxdCEnvVars{EnvVars: make(map[string]interface{}, 0)})
				}
				if h.Out2Var != "" {
					proj.Environments[len(proj.Environments)-1].EnvVars[h.Out2Var] = envs[h.Out2Var]
				}
				if h.Err2Var != "" {
					proj.Environments[len(proj.Environments)-1].EnvVars[h.Err2Var] = envs[h.Err2Var]
				}
			}

			return nil
		}

		// Retrieve list of nodes
		if group != nil {
			nodes = group.Nodes
		} else {
			for _, g := range proj.Groups {
				nodes = append(nodes, g.Nodes...)
			}
		}

		for _, h := range *hooks {

			// Check if hooks must be processed
			if !h.ToProcess(i.FlagsEnabled, i.FlagsDisabled) {
				i.Logger.Debug("Skipped hooks ", h)
				continue
			}

			if h.Commands != nil && len(h.Commands) > 0 {

				for _, cmds := range h.Commands {
					switch h.Node {
					case "", "*":
						for _, node := range nodes {

							// Initialize entrypoint to ensure to set always the
							if node.Entrypoint != nil && len(node.Entrypoint) > 0 {
								executor.Entrypoint = node.Entrypoint
							} else {
								executor.Entrypoint = []string{}
							}

							err := runSingleCmd(&h, node.Name, cmds)
							if err != nil {
								return err
							}
						}

					default:
						err := runSingleCmd(&h, h.Node, cmds)
						if err != nil {
							return err
						}
					}

				}

			}
		}
	}

	return nil
}

func (i *LxdCInstance) ApplyGroup(group *specs.LxdCGroup, proj *specs.LxdCProject, env *specs.LxdCEnvironment, compiler template.LxdCTemplateCompiler) error {

	// Initialize executor
	executor := executor.NewLxdCExecutor(group.Connection,
		i.Config.GetGeneral().LxdConfDir, []string{}, group.Ephemeral,
		i.Config.GetLogging().CmdsOutput)
	err := executor.Setup()
	if err != nil {
		return err
	}

	var syncSourceDir string
	envBaseAbs, err := filepath.Abs(filepath.Dir(env.File))
	if err != nil {
		return err
	}

	// Retrieve pre-group hooks from project
	preGroupHooks := proj.GetHooks4Nodes("pre-group", "host")
	// Retrieve pre-group hooks from group
	preGroupHooks = append(preGroupHooks, group.GetHooks4Nodes("pre-group", "*")...)

	// Run pre-group hooks
	err = i.ProcessHooks(&preGroupHooks, executor, proj, group)
	if err != nil {
		return err
	}

	// Compile group templates
	err = template.CompileGroupFiles(group, compiler, template.CompilerOpts{})
	if err != nil {
		return err
	}

	// TODO: implement parallel creation
	for _, node := range group.Nodes {

		syncSourceDir = ""

		// Initialize entrypoint to ensure to set always the
		if node.Entrypoint != nil && len(node.Entrypoint) > 0 {
			executor.Entrypoint = node.Entrypoint
		} else {
			executor.Entrypoint = []string{}
		}

		isPresent, err := executor.IsPresentContainer(node.Name)
		if err != nil {
			i.Logger.Error("Error on check if container " + node.Name + " is present: " + err.Error())
			return err
		}

		if !isPresent {

			// Retrieve pre-node-creation hooks
			preCreationHooks := i.GetNodeHooks4Event("pre-node-creation", proj, group, &node)
			// Run pre-node-creation hooks
			err = i.ProcessHooks(&preCreationHooks, executor, proj, group)
			if err != nil {
				return err
			}

			profiles := []string{}
			profiles = append(profiles, group.CommonProfiles...)
			profiles = append(profiles, node.Profiles...)

			err := executor.CreateContainer(node.Name, node.ImageSource,
				node.ImageRemoteServer, profiles)
			if err != nil {
				i.Logger.Error("Error on create container " + node.Name + ":" + err.Error())
				return err
			}

			postCreationHooks := i.GetNodeHooks4Event("post-node-creation", proj, group, &node)

			// Run post-node-creation hooks
			err = i.ProcessHooks(&postCreationHooks, executor, proj, group)
			if err != nil {
				return err
			}

		}

		// Retrieve pre-node-sync hooks of the node from project
		preSyncHooks := i.GetNodeHooks4Event("pre-node-sync", proj, group, &node)

		// Run pre-node-sync hooks
		err = i.ProcessHooks(&preSyncHooks, executor, proj, group)
		if err != nil {
			return err
		}

		// Compile node templates
		err = template.CompileNodeFiles(node, compiler, template.CompilerOpts{})
		if err != nil {
			return err
		}

		if len(node.SyncResources) > 0 {
			if node.SourceDir != "" {
				if node.IsSourcePathRelative() {
					syncSourceDir = filepath.Join(envBaseAbs, node.SourceDir)
				} else {
					syncSourceDir = node.SourceDir
				}
			} else {
				// Use env file directory
				syncSourceDir = envBaseAbs
			}

			i.Logger.Debug(i.Logger.Aurora.Bold(
				i.Logger.Aurora.BrightCyan(
					">>> [" + node.Name + "] Using sync source basedir " + syncSourceDir)))

			nResources := len(node.SyncResources)
			i.Logger.InfoC(
				i.Logger.Aurora.Bold(
					i.Logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] Syncing %d resources... - :bus:",
							node.Name, nResources))))

			for idx, resource := range node.SyncResources {

				i.Logger.DebugC(
					i.Logger.Aurora.Italic(
						i.Logger.Aurora.BrightCyan(
							fmt.Sprintf(">>> [%s] %s => %s",
								node.Name, resource.Source, resource.Destination))))

				err = executor.RecursivePushFile(node.Name, filepath.Join(syncSourceDir, resource.Source),
					resource.Destination)
				if err != nil {
					i.Logger.Error("Error on sync " + resource.Source + ": " + err.Error())
					return err
				}

				i.Logger.InfoC(
					i.Logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] - [%2d/%2d] %s - :check_mark:",
							node.Name, idx+1, nResources, resource.Destination)))
			}

		}

		// Retrieve post-node-sync hooks of the node from project
		postSyncHooks := i.GetNodeHooks4Event("post-node-sync", proj, group, &node)

		// Run post-node-sync hooks
		err = i.ProcessHooks(&postSyncHooks, executor, proj, group)
		if err != nil {
			return err
		}

	}

	// Retrieve post-group hooks from project
	postGroupHooks := proj.GetHooks4Nodes("post-group", "host")
	postGroupHooks = append(postGroupHooks, proj.GetHooks4Nodes("post-group", "*")...)

	// Execute post-group hooks
	err = i.ProcessHooks(&postGroupHooks, executor, proj, group)
	if err != nil {
		return err
	}

	return nil
}
