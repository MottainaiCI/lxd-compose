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
package cmd

import (
	"fmt"
	"os"

	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newApplyCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var enabledFlags []string
	var disabledFlags []string
	var envs []string

	var cmd = &cobra.Command{
		Use:   "apply [list-of-projects]",
		Short: "Deploy projects.",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No project selected.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			composer.SetFlagsDisabled(disabledFlags)
			composer.SetFlagsEnabled(enabledFlags)

			projects := args[0:]

			for _, proj := range projects {

				fmt.Println("Apply project " + proj)

				env := composer.GetEnvByProjectName(proj)
				if env == nil {
					fmt.Println("Project " + proj + " not found")
					os.Exit(1)
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

					pObj := env.GetProjectByName(proj)
					pObj.AddEnvironment(evars)
				}

				err = composer.ApplyProject(proj)
				if err != nil {
					fmt.Println("Error on apply project " + proj + ": " + err.Error())
					os.Exit(1)
				}

			}

			fmt.Println("All done.")
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVar(&enabledFlags, "enable-flag", []string{},
		"Run hooks of only specified flags.")
	flags.StringSliceVar(&disabledFlags, "disable-flag", []string{},
		"Disable execution of the hooks with the specified flags.")

	flags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")

	return cmd
}
