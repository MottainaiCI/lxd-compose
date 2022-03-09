/*

Copyright (C) 2020-2022  Daniele Rondina <geaaru@funtoo.org>
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
	template "github.com/MottainaiCI/lxd-compose/pkg/template"

	"github.com/spf13/cobra"
)

func newCompileCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var projects []string
	var sources []string
	var envs []string
	var renderEnvs []string
	var varsFiles []string
	var enabledGroups []string
	var disabledGroups []string
	var cmd = &cobra.Command{
		Use:     "compile",
		Short:   "Compile project templates.",
		Aliases: []string{"co"},
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			prefix, _ := cmd.Flags().GetString("nodes-prefix")

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

			composer.SetNodesPrefix(prefix)

			opts := template.CompilerOpts{
				Sources:        sources,
				GroupsEnabled:  enabledGroups,
				GroupsDisabled: disabledGroups,
			}

			if len(projects) > 0 {
				for _, proj := range projects {

					env := composer.GetEnvByProjectName(proj)
					if env == nil {
						fmt.Println("Project " + proj + " not found")
						os.Exit(1)
					}

					if len(varsFiles) > 0 || len(envs) > 0 {

						pObj := env.GetProjectByName(proj)

						for _, varFile := range varsFiles {
							err := pObj.LoadEnvVarsFile(varFile)
							if err != nil {
								fmt.Println(fmt.Sprintf(
									"Error on load additional envs var file %s: %s",
									varFile, err.Error()))
								os.Exit(1)
							}
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

							pObj.AddEnvironment(evars)
						}
					}

					err := template.CompileAllProjectFiles(env, proj, opts)
					if err != nil {
						fmt.Println("Error on compile files of the project " +
							proj + ":" + err.Error() + "\n")
						os.Exit(1)
					}

				}
			} else {
				for _, env := range *composer.GetEnvironments() {
					for _, proj := range *env.GetProjects() {
						err := template.CompileAllProjectFiles(&env, proj.GetName(), opts)
						if err != nil {
							fmt.Println("Error on compile files of the project " +
								proj.GetName() + ":" + err.Error() + "\n")
							os.Exit(1)
						}
					}
				}
			}

			fmt.Println("Compilation completed!")
		},
	}

	pflags := cmd.Flags()
	pflags.StringSliceVarP(&projects, "project", "p", []string{},
		"Choice the list of the projects to compile. Default: all")
	pflags.StringSliceVarP(&sources, "source-file", "f", []string{},
		"Choice the list of the source file to compile. Default: all")
	pflags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	pflags.StringSliceVar(&varsFiles, "vars-file", []string{},
		"Add additional environments vars file.")
	pflags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	pflags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	pflags.StringSliceVar(&disabledGroups, "disable-group", []string{},
		"Skip selected group from deploy.")
	pflags.StringSliceVar(&enabledGroups, "enable-group", []string{},
		"Apply only selected groups.")

	return cmd
}
