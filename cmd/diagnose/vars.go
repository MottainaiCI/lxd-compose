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
package cmd_diagnose

import (
	"fmt"
	"os"

	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
	"github.com/MottainaiCI/lxd-compose/pkg/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewVarsCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "vars [project]",
		Short: "Dump variables of the project.",
		Args:  cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing project name param")
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

			pName := args[0]

			env := composer.GetEnvByProjectName(pName)
			if env == nil {
				fmt.Println("Project " + pName + " not found")
				os.Exit(1)
			}

			proj := env.GetProjectByName(pName)

			compiler, err := template.NewProjectTemplateCompiler(env, proj)
			if err != nil {
				fmt.Println("Error on initialize compiler: " + err.Error())
				os.Exit(1)
			}

			out, err := yaml.Marshal(*compiler.GetVars())
			if err != nil {
				fmt.Println("Error on convert vars to yaml: " + err.Error())
				os.Exit(1)
			}

			fmt.Println(string(out))

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.Bool("hooks", false, "Execute post-node-sync hooks.")

	return cmd
}
