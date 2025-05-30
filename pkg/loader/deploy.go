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
package loader

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
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
	preProjHooks := proj.GetHooks4Nodes(specs.HookPreProject, []string{"host"})
	postProjHooks := proj.GetHooks4Nodes(specs.HookPostProject, []string{"*", "host"})

	// Execute pre-project hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s] Running %d %s hooks... ", projectName,
		len(preProjHooks), specs.HookPreProject))
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
		"[%s] Running %d %s hooks... ", projectName,
		len(preProjHooks), specs.HookPostProject))
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

						if group == nil && grp == nil {
							return errors.New(fmt.Sprintf(
								"Error on retrieve node information for %s and hook %v",
								node, h))
						}

						if group == nil {
							group = grp
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
				executor.SetP2PMode(i.Config.GetGeneral().P2PMode)

				// NOTE: I don't need to run executor.Setup() for host node.
			}

			if h.Out2Var != "" || h.Err2Var != "" {
				storeVar = true
			} else {
				storeVar = false
			}

			if h.Node == "host" {
				if storeVar {
					res, err = executor.RunHostCommandWithOutput4Var(
						cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint,
					)
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
					res, err = executor.RunCommandWithOutput4Var(
						node, cmds, h.Out2Var, h.Err2Var, &envs, h.Entrypoint,
						h.Uid, h.Gid, h.Cwd,
					)
				} else {
					if i.Config.GetLogging().RuntimeCmdsOutput {
						emitter := executor.GetEmitter()
						res, err = executor.RunCommandWithOutput(
							node, cmds, envs,
							(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStdout(),
							(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStderr(),
							h.Entrypoint, h.Uid, h.Gid, h.Cwd,
						)
					} else {
						res, err = executor.RunCommand(
							node, cmds, envs, h.Entrypoint,
							h.Uid, h.Gid, h.Cwd,
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

	var syncSourceDir string
	envBaseAbs, err := filepath.Abs(filepath.Dir(env.File))
	if err != nil {
		return err
	}

	// Retrieve pre-group hooks from project
	preGroupHooks := proj.GetHooks4Nodes(specs.HookPreGroup, []string{"*", "host"})
	// Retrieve pre-group hooks from group
	preGroupHooks = append(preGroupHooks, group.GetHooks4Nodes(specs.HookPreGroup, []string{"*", "host"})...)

	// Run pre-group hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d %s hooks... ", proj.Name, group.Name,
		len(preGroupHooks), specs.HookPreGroup))
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

	// Initialize executor
	executor := lxd_executor.NewLxdCExecutor(group.Connection,
		i.Config.GetGeneral().LxdConfDir, []string{}, group.Ephemeral,
		i.Config.GetLogging().CmdsOutput,
		i.Config.GetLogging().RuntimeCmdsOutput)
	err = executor.Setup()
	if err != nil {
		return err
	}
	executor.SetP2PMode(i.Config.GetGeneral().P2PMode)

	// Retrieve the list of configured profiles
	instanceProfiles, err := executor.GetProfilesList()
	if err != nil {
		return errors.New(
			fmt.Sprintf("Error on retrieve the list of instance profile of the group %s: %s",
				group.Name, err.Error()))
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

			// Execute the pre-node-creation hooks,
			// create the container and run the post-node-creation
			// hooks.
			err := i.createInstance(
				proj, group, &node,
				executor,
				instanceProfiles,
			)
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

			if i.Upgrade {

				// POST: The instance is already present
				//       but the upgrade flag is enable.

				if isRunning {

					if i.Ask {
						wantUpgrade := helpers.Ask(
							fmt.Sprintf(
								"[%s - %s] Found running node %s, are you sure to proceed with the upgrade? [y/N]: ",
								proj.Name, group.Name, node.GetName(),
							))
						if !wantUpgrade {
							return fmt.Errorf(
								"Upgrade process stopped by user.",
							)
						}
					}

					preNodeUpgradeHooks := i.GetNodeHooks4Event(
						specs.HookPreNodeUpgrade,
						proj, group, &node)

					// Run post-node-creation hooks
					i.Logger.Debug(fmt.Sprintf(
						"[%s - %s] Running %d %s hooks for node %s... ",
						proj.Name, group.Name, len(preNodeUpgradeHooks),
						specs.HookPreNodeUpgrade, node.GetName()))
					err = i.ProcessHooks(&preNodeUpgradeHooks, proj, group, &node)
					if err != nil {
						return err
					}

				} else if i.Ask {
					wantUpgrade := helpers.Ask(
						fmt.Sprintf(
							"[%s - %s] Found stopped node %s, are you sure to proceed with the upgrade of the node? [y/N]: ",
							proj.Name, group.Name, node.GetName(),
						))
					if !wantUpgrade {
						return fmt.Errorf(
							"Upgrade process stopped by user.",
						)
					}
				}

				// POST: The running container is stopped
				//       and destroyed.
				err = executor.DeleteContainer(node.GetName())
				if err != nil {
					i.Logger.Error("Error on destroy container " + node.GetName() +
						": " + err.Error())
					return err
				}

				// Execute the pre-node-creation hooks,
				// create the container and run the post-node-creation
				// hooks.
				err := i.createInstance(
					proj, group, &node,
					executor,
					instanceProfiles,
				)
				if err != nil {
					return err
				}

				postNodeUpgradeHooks := i.GetNodeHooks4Event(
					specs.HookPostNodeUpgrade,
					proj, group, &node)

				// Run post-node-creation hooks
				i.Logger.Debug(fmt.Sprintf(
					"[%s - %s] Running %d %s hooks for node %s... ",
					proj.Name, group.Name, len(postNodeUpgradeHooks),
					specs.HookPostNodeUpgrade, node.GetName()))
				err = i.ProcessHooks(&postNodeUpgradeHooks, proj, group, &node)
				if err != nil {
					return err
				}

			} else {

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

		}

		// Retrieve pre-node-sync hooks of the node from project
		preSyncHooks := i.GetNodeHooks4Event(specs.HookPreNodeSync, proj, group, &node)

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
		postSyncHooks := i.GetNodeHooks4Event(specs.HookPostNodeSync, proj, group, &node)

		// Run post-node-sync hooks
		err = i.ProcessHooks(&postSyncHooks, proj, group, &node)
		if err != nil {
			return err
		}

	}

	// Retrieve post-group hooks from project
	postGroupHooks := proj.GetHooks4Nodes(specs.HookPostGroup, []string{"*", "host"})
	postGroupHooks = append(postGroupHooks, group.GetHooks4Nodes(specs.HookPostGroup, []string{"*", "host"})...)

	// Execute post-group hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d %s hooks... ", proj.Name, group.Name,
		len(postGroupHooks), specs.HookPostGroup))
	err = i.ProcessHooks(&postGroupHooks, proj, group, nil)

	return err
}

func (i *LxdCInstance) createInstance(
	proj *specs.LxdCProject,
	group *specs.LxdCGroup,
	node *specs.LxdCNode,
	executor *lxd_executor.LxdCExecutor,
	instanceProfiles []string) error {

	// Retrieve pre-node-creation hooks
	preCreationHooks := i.GetNodeHooks4Event(specs.HookPreNodeCreation, proj, group, node)
	// Run pre-node-creation hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d %s hooks for node %s... ",
		proj.Name, group.Name, len(preCreationHooks),
		specs.HookPreNodeCreation,
		node.GetName()))
	err := i.ProcessHooks(&preCreationHooks, proj, group, node)
	if err != nil {
		return err
	}

	profiles := []string{}
	profiles = append(profiles, group.CommonProfiles...)
	profiles = append(profiles, node.Profiles...)

	configMap := node.GetLxdConfig(group.GetLxdConfig())

	i.Logger.Debug(fmt.Sprintf("[%s] Using profiles %s",
		node.GetName(), profiles))

	i.Logger.Debug(fmt.Sprintf("[%s] Using config map %s",
		node.GetName(), configMap))

	err = i.validateProfiles(instanceProfiles, profiles)
	if err != nil {
		return err
	}

	err = executor.CreateContainerWithConfig(node.GetName(), node.ImageSource,
		node.ImageRemoteServer, profiles, configMap)
	if err != nil {
		i.Logger.Error("Error on create container " +
			node.GetName() + ":" + err.Error())
		return err
	}

	// Wait ip
	if node.Wait4Ip() > 0 {
		err = executor.WaitIpOfContainer(node.GetName(), node.Wait4Ip())
		if err != nil {
			i.Logger.Error("Something goes wrong on waiting for the ip address: " +
				err.Error())
			return err
		}
	}

	postCreationHooks := i.GetNodeHooks4Event(specs.HookPostNodeCreation, proj, group, node)

	// Run post-node-creation hooks
	i.Logger.Debug(fmt.Sprintf(
		"[%s - %s] Running %d %s hooks for node %s... ",
		proj.Name, group.Name, len(postCreationHooks),
		specs.HookPostNodeCreation, node.GetName()))
	err = i.ProcessHooks(&postCreationHooks, proj, group, node)
	if err != nil {
		return err
	}

	return nil
}

func (i *LxdCInstance) ApplyCommand(c *specs.LxdCCommand, proj *specs.LxdCProject, envs []string, varfiles []string) error {

	if c == nil {
		return errors.New("Invalid command")
	}

	if proj == nil {
		return errors.New("Invalid project")
	}

	env := i.GetEnvByProjectName(proj.GetName())
	if env == nil {
		return errors.New(fmt.Sprintf("No environment found for project " + proj.GetName()))
	}

	envBaseDir, err := filepath.Abs(filepath.Dir(env.File))
	if err != nil {
		return err
	}

	// Load envs from commands.
	if len(c.VarFiles) > 0 {
		for _, varFile := range c.VarFiles {

			envs, err := i.loadEnvFile(envBaseDir, varFile, proj)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on load additional envs var file %s: %s",
						varFile, err.Error()),
				)
			}

			proj.AddEnvironment(envs)

		}
	}

	if len(c.Envs.EnvVars) > 0 {
		proj.AddEnvironment(&c.Envs)
	}

	if len(c.IncludeHooksFiles) > 0 {

		for _, hfile := range c.IncludeHooksFiles {

			// Load project included hooks
			hf := path.Join(envBaseDir, hfile)
			hooks, err := i.getHooks(hfile, hf, proj)
			if err != nil {
				return err
			}

			proj.AddHooks(hooks)
		}
	}

	if len(envs) > 0 {
		evars := specs.NewEnvVars()
		for _, e := range envs {
			err := evars.AddKVAggregated(e)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on elaborate var string %s: %s",
						e, err.Error(),
					))
			}
		}

		proj.AddEnvironment(evars)
	}

	i.SetFlagsDisabled(c.DisableFlags)
	i.SetFlagsEnabled(c.EnableFlags)
	i.SetGroupsDisabled(c.DisableGroups)
	i.SetGroupsEnabled(c.EnableGroups)
	i.SetSkipSync(c.SkipSync)
	i.SetNodesPrefix(c.NodesPrefix)

	return nil
}

func (i *LxdCInstance) validateProfiles(instanceProfiles, requiredProfiles []string) error {
	mProf := make(map[string]bool, 0)

	for _, p := range instanceProfiles {
		mProf[p] = true
	}

	for _, p := range requiredProfiles {
		if _, ok := mProf[p]; !ok {
			return errors.New(fmt.Sprintf("Profile %s not available on target instance.", p))
		}
	}

	return nil
}
