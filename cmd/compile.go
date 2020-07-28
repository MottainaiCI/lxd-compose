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
	template "github.com/MottainaiCI/lxd-compose/pkg/template"

	"github.com/spf13/cobra"
)

func newCompileCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var projects []string
	var sources []string
	var cmd = &cobra.Command{
		Use:   "compile",
		Short: "Compile project templates.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			opts := template.CompilerOpts{
				Sources: sources,
			}

			if len(projects) > 0 {
				for _, proj := range projects {

					env := composer.GetEnvByProjectName(proj)
					if env == nil {
						fmt.Println("Project " + proj + " not found")
						os.Exit(1)
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

	return cmd
}
