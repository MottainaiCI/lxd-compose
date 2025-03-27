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
	"encoding/json"
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "list of nodes available to the specified endpoint.",
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			if confdir == "" {
				// Using lxd-compose config option if available
				confdir = config.GetGeneral().LxdConfDir
			}

			executor := executor.NewLxdCExecutor(endpoint, confdir, nil, true,
				config.GetLogging().CmdsOutput,
				config.GetLogging().RuntimeCmdsOutput)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
				os.Exit(1)
			}

			list, err := executor.GetContainerList()
			if err != nil {
				fmt.Println("Error on retrieve container list: " + err.Error() + "\n")
				os.Exit(1)
			}

			if search != "" {
				list = helpers.RegexEntry(search, list)
			}

			if jsonOutput {
				data, _ := json.Marshal(list)
				fmt.Println(string(data))

			} else {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})

				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Node Name", "Project Name", "Group Name",
				})

				for _, n := range list {

					pName := ""
					gName := ""

					_, proj, group, _ := composer.GetEntitiesByNodeName(n)

					if proj != nil {
						pName = proj.GetName()
						gName = group.GetName()
					}

					table.Append([]string{
						n,
						pName,
						gName,
					})

				}

				table.Render()
			}
		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.Bool("json", false, "JSON output")
	pflags.StringP("search", "s", "", "Regex filter to use with node name.")

	return cmd
}
