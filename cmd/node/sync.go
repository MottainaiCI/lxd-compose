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
package cmd_node

import (
	"fmt"
	"os"
	"path/filepath"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewSyncCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "sync [node]",
		Short:   "Sync node files.",
		Aliases: []string{"s"},
		Args:    cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing node name param")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			syncSourceDir := ""
			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			syncPostCmds, _ := cmd.Flags().GetBool("hooks")
			prefix, _ := cmd.Flags().GetString("nodes-prefix")

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

			env, proj, grp, nodeConf := composer.GetEntitiesByNodeName(node)
			if env == nil && prefix != "" {
				// Check if i find the node with prefix
				env, proj, grp, nodeConf = composer.GetEntitiesByNodeName(
					fmt.Sprintf("%s-%s", prefix, node))
			}

			if env == nil {
				fmt.Println("Node not found")
				os.Exit(1)
			}

			if endpoint == "" {
				endpoint = grp.Connection
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

			if len(nodeConf.SyncResources) == 0 {
				fmt.Println("No resources to sync available.")
				os.Exit(0)
			}

			envBaseAbs, err := filepath.Abs(filepath.Dir(env.File))
			if err != nil {
				fmt.Println("Error on retrieve environment base path: " + err.Error())
				os.Exit(1)
			}

			if nodeConf.SourceDir != "" {
				if nodeConf.IsSourcePathRelative() {
					syncSourceDir = filepath.Join(envBaseAbs, nodeConf.SourceDir)
				} else {
					syncSourceDir = nodeConf.SourceDir
				}
			} else {
				// Use env file directory
				syncSourceDir = envBaseAbs
			}

			fmt.Println("Using sync source basedir ", syncSourceDir)

			for _, resource := range nodeConf.SyncResources {

				fmt.Println("Syncing resource " + resource.Source + " => " + resource.Destination)

				err = executor.RecursivePushFile(node, filepath.Join(syncSourceDir, resource.Source),
					resource.Destination)
				if err != nil {
					fmt.Println("Error on sync " + resource.Source + ": " + err.Error())
					os.Exit(1)
				}
			}

			if syncPostCmds {
				hooks := composer.GetNodeHooks4Event("post-node-sync", proj, grp, nodeConf)
				err := composer.ProcessHooks(&hooks, proj, grp, nodeConf)
				if err != nil {
					fmt.Println("Error " + err.Error())
					os.Exit(1)
				}

			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.Bool("hooks", false, "Execute post-node-sync hooks.")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")

	return cmd
}
