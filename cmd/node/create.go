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

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewCreateCommand(config *specs.LxdComposeConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "create node1 node2",
		Short: "Create one or more nodes.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			postCreationHooks, _ := cmd.Flags().GetBool("hooks")
			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
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

				executor := executor.NewLxdCExecutor(endpoint, confdir,
					nodeConf.Entrypoint, grp.Ephemeral, config.GetLogging().CmdsOutput)
				err = executor.Setup()
				if err != nil {
					fmt.Println("Error on setup executor:" + err.Error() + "\n")
					os.Exit(1)
				}

				// Create container
				fmt.Println("Creating ... ", n)

				profiles := []string{}
				profiles = append(profiles, grp.CommonProfiles...)
				profiles = append(profiles, nodeConf.Profiles...)

				err := executor.CreateContainer(n, nodeConf.ImageSource,
					nodeConf.ImageRemoteServer, profiles)
				if err != nil {
					fmt.Println("Error on create container "+n+":", err.Error())
					os.Exit(1)
				}

				envs := proj.GetEnvsMap()
				if _, ok := envs["HOME"]; !ok {
					envs["HOME"] = "/"
				}

				if postCreationHooks {
					hooks := composer.GetNodeHooks4Event("post-node-creation", proj, grp, nodeConf)
					err := composer.ProcessHooks(&hooks, executor, proj, grp)
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

	return cmd
}
