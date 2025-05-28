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
package cmd_profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		Aliases: []string{"l", "li"},
		Short:   "list of LXD profiles available to endpoint.",
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

			list, err := executor.GetProfilesList()
			if err != nil {
				fmt.Println("Error on retrieve profile list: " + err.Error() + "\n")
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
				table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"Profile Name", "Description", "Environments Mapped",
				})

				for _, c := range list {

					mappedEnvs := ""
					profileDescr := ""
					envs := composer.GetEnvsUsingProfile(c)
					if len(envs) > 0 {
						for idx, e := range envs {
							profile, _ := e.GetProfile(c)
							profileDescr = profile.Description
							if idx > 0 {
								mappedEnvs += "\n" + filepath.Base(e.File)
							} else {
								mappedEnvs += filepath.Base(e.File)
							}
						}
					}

					table.Append([]string{
						c,
						profileDescr,
						mappedEnvs,
					})

				}

				table.Render()
			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.Bool("json", false, "JSON output")
	pflags.StringP("search", "s", "", "Regex filter to use with profile name.")

	return cmd
}
