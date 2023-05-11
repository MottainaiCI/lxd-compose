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
package cmd_node

import (
	"fmt"
	"os"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewCreateCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var enabledFlags []string
	var disabledFlags []string
	var envs []string

	var cmd = &cobra.Command{
		Use:     "create node1 node2",
		Aliases: []string{"c"},
		Short:   "Create one or more nodes.",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			postCreationHooks, _ := cmd.Flags().GetBool("hooks")
			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			prefix, _ := cmd.Flags().GetString("nodes-prefix")
			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			composer.SetFlagsDisabled(disabledFlags)
			composer.SetFlagsEnabled(enabledFlags)
			composer.SetNodesPrefix(prefix)

			if confdir == "" {
				confdir = config.GetGeneral().LxdConfDir
			}

			nodes := args[0:]

			for _, n := range nodes {

				env, proj, grp, nodeConf := composer.GetEntitiesByNodeName(n)
				if env == nil {
					fmt.Println("Skipped node", n)
					continue
				}

				if endpoint == "" {
					endpoint = grp.Connection
				}

				if len(envs) > 0 {

					evars := specs.NewEnvVars()
					for _, e := range envs {
						err := evars.AddKVAggregated(e)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					}

					proj.AddEnvironment(evars)
				}

				executor := lxd_executor.NewLxdCExecutor(endpoint, confdir,
					nodeConf.Entrypoint, grp.Ephemeral,
					config.GetLogging().CmdsOutput,
					config.GetLogging().RuntimeCmdsOutput)
				err = executor.Setup()
				if err != nil {
					fmt.Println("Error on setup executor:" + err.Error() + "\n")
					os.Exit(1)
				}

				// Set p2p mode
				executor.SetP2PMode(config.GetGeneral().P2PMode)

				// Create container
				fmt.Println("Creating ... ", n)

				profiles := []string{}
				profiles = append(profiles, grp.CommonProfiles...)
				profiles = append(profiles, nodeConf.Profiles...)

				configMap := nodeConf.GetLxdConfig(grp.GetLxdConfig())

				err := executor.CreateContainerWithConfig(n, nodeConf.ImageSource,
					nodeConf.ImageRemoteServer, profiles, configMap)
				if err != nil {
					fmt.Println("Error on create container "+n+":", err.Error())
					os.Exit(1)
				}

				envs, err := proj.GetEnvsMap()
				if err != nil {
					fmt.Println("Error on convert variables in envs:" + err.Error() + "\n")
					os.Exit(1)
				}
				if _, ok := envs["HOME"]; !ok {
					envs["HOME"] = "/"
				}

				if postCreationHooks {
					hooks := composer.GetNodeHooks4Event("post-node-creation", proj, grp, nodeConf)
					err := composer.ProcessHooks(&hooks, proj, grp, nodeConf)
					if err != nil {
						fmt.Println("Error " + err.Error())
						os.Exit(1)
					}
				}
			}
		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.Bool("hooks", false, "Execute post-node-creation hooks")
	pflags.StringSliceVar(&enabledFlags, "enable-flag", []string{},
		"Run hooks of only specified flags.")
	pflags.StringSliceVar(&disabledFlags, "disable-flag", []string{},
		"Disable execution of the hooks with the specified flags.")
	pflags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")

	return cmd
}
