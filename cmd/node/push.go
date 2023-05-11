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

func NewPushCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "push [node] [opts]",
		Aliases: []string{"p", "pu"},
		Short:   "Push files to node.",
		Args:    cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing node name param")
				os.Exit(1)
			}

			sourcePath, _ := cmd.Flags().GetString("from")
			targetPath, _ := cmd.Flags().GetString("to")

			if sourcePath == "" || targetPath == "" {
				fmt.Println("Missing mandatory --to or --from options.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			prefix, _ := cmd.Flags().GetString("nodes-prefix")
			sourcePath, _ := cmd.Flags().GetString("from")
			targetPath, _ := cmd.Flags().GetString("to")

			// Create Instance
			composer := loader.NewLxdCInstance(config)
			endpoint, _ := cmd.Flags().GetString("endpoint")

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			if confdir == "" {
				// Using lxd-compose config option if available
				confdir = config.GetGeneral().LxdConfDir
			}

			composer.SetNodesPrefix(prefix)

			node := args[0]
			entrypoint := []string{}

			env, _, grp, nodeConf := composer.GetEntitiesByNodeName(node)
			if env == nil && prefix != "" {
				// Check if i find the node with prefix
				env, _, grp, nodeConf = composer.GetEntitiesByNodeName(
					fmt.Sprintf("%s-%s", prefix, node))
			}

			if env != nil && nodeConf != nil {
				if endpoint == "" && grp != nil {
					endpoint = grp.Connection
				}
				entrypoint = nodeConf.Entrypoint
			}

			if endpoint == "" && grp == nil {
				fmt.Println("Node not found and endpoint argument missing.")
				os.Exit(1)
			}

			executor := lxd_executor.NewLxdCExecutor(endpoint, confdir,
				entrypoint, false,
				config.GetLogging().CmdsOutput,
				config.GetLogging().RuntimeCmdsOutput)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
				os.Exit(1)
			}

			err = executor.RecursivePushFile(node, sourcePath, targetPath)
			if err != nil {
				fmt.Println("Error on push " + sourcePath + ": " + err.Error())
				os.Exit(1)
			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	pflags.String("from", "", "Source host path.")
	pflags.String("to", "", "Target container path")

	return cmd
}
