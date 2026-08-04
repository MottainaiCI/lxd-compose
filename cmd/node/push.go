/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
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
			connType, _ := cmd.Flags().GetString("connection-type")

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
					connType = grp.ConnectionType
				}
				entrypoint = nodeConf.Entrypoint
			}

			if endpoint == "" && grp == nil {
				fmt.Println("Node not found and endpoint argument missing.")
				os.Exit(1)
			}

			executor := lxd_executor.NewLxdCExecutor(
				connType, endpoint, confdir,
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
	pflags.String("connection-type", "incus", "Set connection type.")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	pflags.String("from", "", "Source host path.")
	pflags.String("to", "", "Target container path")

	return cmd
}
