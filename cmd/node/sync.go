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
	"path/filepath"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewSyncCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync [node]",
		Short: "Sync node files.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			syncSourceDir := ""
			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			syncPostCmds, _ := cmd.Flags().GetBool("hooks")

			// Create Instance
			composer := loader.NewLxdCInstance(config)
			endpoint, _ := cmd.Flags().GetString("endpoint")

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			node := args[0]

			env, proj, grp, nodeConf := composer.GetEntitiesByNodeName(node)
			if env == nil {
				fmt.Println("Node not found")
				os.Exit(1)
			}

			if endpoint == "" {
				endpoint = grp.Connection
			}

			executor := executor.NewLxdCExecutor(endpoint, confdir, nodeConf.Entrypoint, grp.Ephemeral)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
				os.Exit(1)
			}

			if len(nodeConf.SyncResources) == 0 {
				fmt.Println("No resources to sync available.")
				os.Exit(0)
			}

			if nodeConf.SourceDir != "" {
				if nodeConf.IsSourcePathRelative() {
					syncSourceDir = filepath.Join(filepath.Dir(env.File), nodeConf.SourceDir)
				} else {
					syncSourceDir = nodeConf.SourceDir
				}
			} else {
				// Use env file directory
				syncSourceDir = filepath.Dir(env.File)
			}

			fmt.Println("Using sync source basedir ", syncSourceDir)

			for _, resource := range nodeConf.SyncResources {

				fmt.Println("Syncing resource " + resource.Source + " => " + resource.Destination)

				err = executor.RecursivePushFile(node, filepath.Join(syncSourceDir, resource.Source),
					filepath.Dir(resource.Destination)+"/")
				if err != nil {
					fmt.Println("Error on sync " + resource.Source + ": " + err.Error())
					os.Exit(1)
				}
			}

			if syncPostCmds {
				hooks := composer.GetNodeHooks4Event("post-node-sync", proj, grp, nodeConf)
				err := composer.ProcessHooks(&hooks, executor, proj, grp)
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

	return cmd
}
