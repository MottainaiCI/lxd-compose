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
package cmd_storage

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
		Use:     "list <project>",
		Aliases: []string{"l"},
		Short:   "List the definitions of the available storages defined in the project.",

		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No project selected.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			project := args[0]
			env := composer.GetEnvByProjectName(project)
			if env == nil {
				fmt.Println("Project not found")
				os.Exit(1)
			}

			storages := *env.GetStorages()

			if search != "" {
				nstorages := []specs.LxdCStorage{}

				for _, s := range storages {
					res := helpers.RegexEntry(search, []string{s.GetName()})
					if len(res) > 0 {
						nstorages = append(nstorages, s)
					}
				}

				storages = nstorages
			}

			if jsonOutput {

				data, _ := json.Marshal(storages)
				fmt.Println(string(data))
			} else {

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Storages", "Driver", "Documentation",
				})
				table.SetAutoWrapText(false)

				for _, s := range storages {
					table.Append([]string{
						s.GetName(),
						s.GetDriver(),
						s.GetDocumentation(),
					})
				}
				table.Render()
			}
		},
	}

	pflags := cmd.Flags()
	pflags.Bool("json", false, "JSON output")
	pflags.StringP("search", "s", "", "Regex filter to use with storage name.")

	return cmd
}
