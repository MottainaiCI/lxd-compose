/*
Copyright (C) 2020-2022  Daniele Rondina <geaaru@sabayonlinux.org>
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

func newDestroyCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string
	var envs []string
	var enabledGroups []string
	var disabledGroups []string

	var cmd = &cobra.Command{
		Use:     "destroy [list-of-projects]",
		Short:   "Destroy projects.",
		Aliases: []string{"d"},
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No project selected.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// We need set this before loading phase
			err := config.SetRenderEnvs(renderEnvs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			prefix, _ := cmd.Flags().GetString("nodes-prefix")

			composer.SetNodesPrefix(prefix)
			composer.SetGroupsDisabled(disabledGroups)
			composer.SetGroupsEnabled(enabledGroups)

			projects := args[0:]

			evars := specs.NewEnvVars()
			if len(envs) > 0 {
				for _, e := range envs {
					err := evars.AddKVAggregated(e)
					if err != nil {
						fmt.Println(
							fmt.Sprintf(
								"Error on elaborate var string %s: %s",
								e, err.Error(),
							))
						os.Exit(1)
					}
				}
			}

			for _, proj := range projects {

				env := composer.GetEnvByProjectName(proj)
				if env == nil {
					fmt.Println("Project " + proj + " not found")
					os.Exit(1)
				}

				if len(envs) > 0 {
					p := env.GetProjectByName(proj)
					p.AddEnvironment(evars)
				}

				err = composer.DestroyProject(proj)
				if err != nil {
					fmt.Println("Error on destroy project " + proj + ": " + err.Error())
					os.Exit(1)
				}

			}

			fmt.Println("All done.")
		},
	}

	flags := cmd.Flags()
	flags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	flags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	flags.StringSliceVar(&disabledGroups, "disable-group", []string{},
		"Skip selected group from deploy.")
	flags.StringSliceVar(&enabledGroups, "enable-group", []string{},
		"Apply only selected groups.")

	return cmd
}
