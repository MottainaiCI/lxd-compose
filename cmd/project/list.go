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
package cmd_diagnose

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "li"},
		Short:   "List all loaded projects.",
		Run: func(cmd *cobra.Command, args []string) {
			projects := []specs.LxdCProjectSanitized{}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)
			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			for _, e := range *composer.GetEnvironments() {
				for _, p := range *e.GetProjects() {
					if search != "" {
						res := helpers.RegexEntry(search, []string{p.GetName()})
						if len(res) > 0 {
							projects = append(projects, *p.Sanitize())
						}
					} else {
						projects = append(projects, *p.Sanitize())
					}
				}
			}

			if jsonOutput {

				data, err := json.Marshal(projects)
				if err != nil {
					fmt.Println("Error on decode projects ", err.Error())
					os.Exit(1)
				}
				fmt.Println(string(data))

			} else {

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Project Name", "Description", "# Groups",
				})
				table.SetAutoWrapText(false)

				for _, p := range projects {

					table.Append([]string{
						p.GetName(),
						p.GetDescription(),
						fmt.Sprintf("%d", len(*p.GetGroups())),
					})
				}

				table.Render()
			}

		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")
	flags.StringP("search", "s", "", "Regex filter to use with network name.")

	return cmd
}
