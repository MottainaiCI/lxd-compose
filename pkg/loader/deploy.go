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
package loader

import (
	"errors"
	"fmt"
	"path/filepath"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
	"github.com/MottainaiCI/lxd-compose/pkg/template"
)

func (i *LxdCInstance) GetNodeHooks4Event(event string, proj *specs.LxdCProject, group *specs.LxdCGroup, node *specs.LxdCNode) []specs.LxdCHook {

	// Retrieve project hooks
	projHooks := proj.GetHooks4Nodes(event, []string{"*"})
	projHooks = specs.FilterHooks4Node(&projHooks, []string{node.GetName(), "host"})

	// Retrieve group hooks
	groupHooks := group.GetHooks4Nodes(event, []string{"*"})
	groupHooks = specs.FilterHooks4Node(&groupHooks, []string{node.GetName(), "host"})

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

	if i.NodesPrefix != "" {
		proj.SetNodesPrefix(i.NodesPrefix)
	}

	// Get only host hooks. All other hooks are handled by group and node.
	preProjHooks := proj.GetHooks4Nodes("pre-project", []string{"host"})
	postProjHooks := proj.GetHooks4Nodes("post-project", []string{"*", "host"})

	// Execute pre-project hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s] Running %d pre-project hooks... ", projectName, len(preProjHooks)))
	err := i.ProcessHooks(&preProjHooks, proj, nil, nil)
	if err != nil {
		return err
	}

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

		if !grp.ToProcess(i.GroupsEnabled, i.GroupsDisabled) {
			i.Logger.Debug("Skipped group ", grp.Name)
			continue
		}

		err := i.ApplyGroup(&grp, proj, env, compiler)
		if err != nil {
			return err
		}

	}

	// Execute post-project hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s] Running %d post-project hooks... ", projectName, len(preProjHooks)))
	err = i.ProcessHooks(&postProjHooks, proj, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (i *LxdCInstance) ProcessHooks(hooks *[]specs.LxdCHook, proj *specs.LxdCProject, group *specs.LxdCGroup, targetNode *specs.LxdCNode) error {
	var res int
	nodes := []specs.LxdCNode{}
	storeVar := false

	executorMap := make(map[string]*lxd_executor.LxdCExecutor, 0)

	if len(*hooks) > 0 {

		runSingleCmd := func(h *specs.LxdCHook, node, cmds string) error {
			var executor *lxd_executor.LxdCExecutor

			envs, err := proj.GetEnvsMap()
			if err != nil {
				return err
			}
			if _, ok := envs["HOME"]; !ok {
				envs["HOME"] = "/"
			}

			if node != "host" {

				var grp *specs.LxdCGroup = nil
				var nodeEntity *specs.LxdCNode = nil

				_, _, grp, nodeEntity = i.GetEntitiesByNodeName(node)
				if nodeEntity == nil && i.NodesPrefix != "" {
					// Trying to search node with prefix
					_, _, grp, nodeEntity = i.GetEntitiesByNodeName(
						fmt.Sprintf("%s-%s", i.NodesPrefix, node))

					if nodeEntity != nil {
						node = fmt.Sprintf("%s-%s", i.NodesPrefix, node)
					}
				}

				if nodeEntity != nil {
					json, err := nodeEntity.ToJson()
					if err != nil {
						return err
					}
					envs["node"] = json

					if nodeEntity.Labels != nil && len(nodeEntity.Labels) > 0 {
						for k, v := range nodeEntity.Labels {
							envs[k] = v
						}
					}

					if _, ok := executorMap[node]; !ok {
						// Initialize executor
						executor = lxd_executor.NewLxdCExecutor(grp.Connection,
							i.Config.GetGeneral().LxdConfDir, []string{}, grp.Ephemeral,
							i.Config.GetLogging().CmdsOutput,
							i.Config.GetLogging().RuntimeCmdsOutput)
						err := executor.Setup()
						if err != nil {
							return err
						}

						executor.SetP2PMode(i.Config.GetGeneral().P2PMode)
						executorMap[node] = executor
					} else {

						if group == nil {
							return errors.New(fmt.Sprintf(
								"Error on retrieve node information for %s and hook %s",
								node, h))
						}

						executor = lxd_executor.NewLxdCExecutor(group.Connection,
							i.Config.GetGeneral().LxdConfDir, []string{}, group.Ephemeral,
							i.Config.GetLogging().CmdsOutput,
							i.Config.GetLogging().RuntimeCmdsOutput)
						err := executor.Setup()
						if err != nil {
							return err
						}
						executor.SetP2PMode(i.Config.GetGeneral().P2PMode)
					}

					// Initialize entrypoint to ensure to set always the
					if nodeEntity.Entrypoint != nil && len(nodeEntity.Entrypoint) > 0 {
						executor.Entrypoint = nodeEntity.Entrypoint
					} else {
						executor.Entrypoint = []string{}
					}

				} else {
					executor = executorMap[node]
				}

			} else {
				connection := "local"
				ephemeral := true

				if group != nil {
					connection = group.Connection
					ephemeral = group.Ephemeral
				}
				// Initialize executor with local LXD connection
				executor = lxd_executor.NewLxdCExecutor(connection,
					i.Config.GetGeneral().LxdConfDir, []string{}, ephemeral,
					i.Config.GetLogging().CmdsOutput,
					i.Config.GetLogging().RuntimeCmdsOutput)
				err := executor.Setup()
				if err != nil {
					return err
				}

				executor.SetP2PMode(i.Config.GetGeneral().P2PMode)
			}

			if h.Out2Var != "" || h.Err2Var != "" {
				storeVar = true
			} else {
				storeVar = false
			}

			if h.Node == "host" {
				if storeVar {
					res, err = executor.RunHostCommandWithOutput4Var(cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint)
				} else {
					if i.Config.GetLogging().RuntimeCmdsOutput {
						emitter := executor.GetEmitter()
						res, err = executor.RunHostCommandWithOutput(
							cmds, envs,
							(emitter.(*lxd_executor.LxdCEmitter)).GetHostWriterStdout(),
							(emitter.(*lxd_executor.LxdCEmitter)).GetHostWriterStderr(),
							h.Entrypoint,
						)
					} else {
						res, err = executor.RunHostCommand(cmds, envs, h.Entrypoint)
					}
				}
			} else {

				if storeVar {
					res, err = executor.RunCommandWithOutput4Var(node, cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint)
				} else {
					if i.Config.GetLogging().RuntimeCmdsOutput {
						emitter := executor.GetEmitter()
						res, err = executor.RunCommandWithOutput(
							node, cmds, envs,
							(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStdout(),
							(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStderr(),
							h.Entrypoint)
					} else {
						res, err = executor.RunCommand(
							node, cmds, envs, h.Entrypoint,
						)
					}
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
						if targetNode != nil {
							err := runSingleCmd(&h, targetNode.GetName(), cmds)
							if err != nil {
								return err
							}
						} else {
							for _, node := range nodes {
								err := runSingleCmd(&h, node.GetName(), cmds)
								if err != nil {
									return err
								}
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
	executor := lxd_executor.NewLxdCExecutor(group.Connection,
		i.Config.GetGeneral().LxdConfDir, []string{}, group.Ephemeral,
		i.Config.GetLogging().CmdsOutput,
		i.Config.GetLogging().RuntimeCmdsOutput)
	err := executor.Setup()
	if err != nil {
		return err
	}
	executor.SetP2PMode(i.Config.GetGeneral().P2PMode)

	var syncSourceDir string
	envBaseAbs, err := filepath.Abs(filepath.Dir(env.File))
	if err != nil {
		return err
	}

	// Retrieve pre-group hooks from project
	preGroupHooks := proj.GetHooks4Nodes("pre-group", []string{"*", "host"})
	// Retrieve pre-group hooks from group
	preGroupHooks = append(preGroupHooks, group.GetHooks4Nodes("pre-group", []string{"*", "host"})...)

	// Run pre-group hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d pre-group hooks... ", proj.Name, group.Name, len(preGroupHooks)))
	err = i.ProcessHooks(&preGroupHooks, proj, group, nil)
	if err != nil {
		return err
	}

	// We need reload variables updated from out2var/err2var hooks.
	compiler.InitVars()

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

		isPresent, err := executor.IsPresentContainer(node.GetName())
		if err != nil {
			i.Logger.Error("Error on check if container " +
				node.GetName() + " is present: " + err.Error())
			return err
		}

		i.Logger.Debug(fmt.Sprintf(
			"[%s - %s] Node %s is present: %v.",
			proj.Name, group.Name, node.GetName(), isPresent))

		if !isPresent {

			// Retrieve pre-node-creation hooks
			preCreationHooks := i.GetNodeHooks4Event("pre-node-creation", proj, group, &node)
			// Run pre-node-creation hooks
			i.Logger.Debug(fmt.Sprintf(
				"[%s - %s] Running %d pre-node-creation hooks for node %s... ",
				proj.Name, group.Name, len(preCreationHooks), node.GetName()))
			err = i.ProcessHooks(&preCreationHooks, proj, group, &node)
			if err != nil {
				return err
			}

			profiles := []string{}
			profiles = append(profiles, group.CommonProfiles...)
			profiles = append(profiles, node.Profiles...)
			i.Logger.Debug(fmt.Sprintf("[%s] Using profiles %s",
				node.GetName(), profiles))

			err := executor.CreateContainer(node.GetName(), node.ImageSource,
				node.ImageRemoteServer, profiles)
			if err != nil {
				i.Logger.Error("Error on create container " +
					node.GetName() + ":" + err.Error())
				return err
			}

			postCreationHooks := i.GetNodeHooks4Event("post-node-creation", proj, group, &node)

			// Run post-node-creation hooks
			i.Logger.Debug(fmt.Sprintf(
				"[%s - %s] Running %d post-node-creation hooks for node %s... ",
				proj.Name, group.Name, len(postCreationHooks), node.GetName()))
			err = i.ProcessHooks(&postCreationHooks, proj, group, &node)
			if err != nil {
				return err
			}

		} else {
			isRunning, err := executor.IsRunningContainer(node.GetName())
			if err != nil {
				i.Logger.Error(
					fmt.Sprintf("Error on check if container %s is running: %s",
						node.GetName(), err.Error()))
				return err
			}
			if !isRunning {
				// Run post-node-creation hooks
				i.Logger.Debug(fmt.Sprintf(
					"[%s - %s] Node %s is already present but not running. I'm starting it.",
					proj.Name, group.Name, node.GetName()))

				err = executor.StartContainer(node.GetName())
				if err != nil {
					i.Logger.Error(
						fmt.Sprintf("Error on start container %s: %s",
							node.GetName(), err.Error()))
					return err
				}
			}

		}

		// Retrieve pre-node-sync hooks of the node from project
		preSyncHooks := i.GetNodeHooks4Event("pre-node-sync", proj, group, &node)

		// Run pre-node-sync hooks
		err = i.ProcessHooks(&preSyncHooks, proj, group, &node)
		if err != nil {
			return err
		}

		// We need reload variables updated from out2var/err2var hooks.
		compiler.InitVars()

		// Compile node templates
		err = template.CompileNodeFiles(node, compiler, template.CompilerOpts{})
		if err != nil {
			return err
		}

		if len(node.SyncResources) > 0 && !i.SkipSync {
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
					">>> [" + node.GetName() + "] Using sync source basedir " +
						syncSourceDir)))

			nResources := len(node.SyncResources)
			i.Logger.InfoC(
				i.Logger.Aurora.Bold(
					i.Logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] Syncing %d resources... - :bus:",
							node.GetName(), nResources))))

			for idx, resource := range node.SyncResources {

				var sourcePath string

				if filepath.IsAbs(resource.Source) {
					sourcePath = resource.Source
				} else {
					sourcePath = filepath.Join(syncSourceDir, resource.Source)
				}

				i.Logger.DebugC(
					i.Logger.Aurora.Italic(
						i.Logger.Aurora.BrightCyan(
							fmt.Sprintf(">>> [%s] %s => %s",
								node.GetName(), resource.Source,
								resource.Destination))))

				err = executor.RecursivePushFile(node.GetName(),
					sourcePath, resource.Destination)
				if err != nil {
					i.Logger.Debug("Error on sync from sourcePath " + sourcePath +
						" to dest " + resource.Destination)
					i.Logger.Error("Error on sync " + resource.Source + ": " + err.Error())
					return err
				}

				i.Logger.InfoC(
					i.Logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] - [%2d/%2d] %s - :check_mark:",
							node.GetName(), idx+1, nResources, resource.Destination)))
			}

		}

		// Retrieve post-node-sync hooks of the node from project
		postSyncHooks := i.GetNodeHooks4Event("post-node-sync", proj, group, &node)

		// Run post-node-sync hooks
		err = i.ProcessHooks(&postSyncHooks, proj, group, &node)
		if err != nil {
			return err
		}

	}

	// Retrieve post-group hooks from project
	postGroupHooks := proj.GetHooks4Nodes("post-group", []string{"*", "host"})
	postGroupHooks = append(postGroupHooks, group.GetHooks4Nodes("post-group", []string{"*", "host"})...)

	// Execute post-group hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d post-group hooks... ", proj.Name, group.Name, len(postGroupHooks)))
	err = i.ProcessHooks(&postGroupHooks, proj, group, nil)

	return err
}
