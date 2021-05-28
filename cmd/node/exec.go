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
	"strings"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewExecCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var envs []string

	var cmd = &cobra.Command{
		Use:     "exec node [command]",
		Aliases: []string{"e", "exec"},
		Short:   "Execute a command to a node or a list of nodes.",
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
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
			commands := args[1:]

			env, proj, grp, nodeConf := composer.GetEntitiesByNodeName(node)
			if env == nil {
				fmt.Println("Node not found")
				os.Exit(1)
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

			executor := lxd_executor.NewLxdCExecutor(endpoint, confdir, nodeConf.Entrypoint,
				grp.Ephemeral, config.GetLogging().CmdsOutput,
				config.GetLogging().RuntimeCmdsOutput)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
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

			emitter := executor.GetEmitter()
			res, err := executor.RunCommandWithOutput(
				node, strings.Join(commands, " "), envs,
				(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStdout(),
				(emitter.(*lxd_executor.LxdCEmitter)).GetLxdWriterStderr(),
				[]string{},
			)

			os.Exit(res)
		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")

	return cmd
}
