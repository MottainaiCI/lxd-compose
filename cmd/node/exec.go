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

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewExecCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "exec node [command]",
		Short: "Execute a command to a node or a list of nodes.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			node := args[0]
			commands := args[1:]

			fmt.Println("node ", node)
			fmt.Println("command ", commands)

			env, _, _ := composer.GetEnvProjGroupByNodeName(node)
			if env == nil {
				fmt.Println("Node not found")
				os.Exit(1)
			}

			executor := executor.NewLxdCExecutor("", "", []string{}, true)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
				os.Exit(1)
			}
			res, err := executor.RunCommand(node, strings.Join(commands, " "),
				map[string]string{
					"HOME": "/",
				})

			os.Exit(res)
		},
	}

	return cmd
}
