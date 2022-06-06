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

	yamlgo "github.com/ghodss/yaml"
	"github.com/icza/dyno"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewVarsCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string
	var envs []string

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

			jsonFormat, _ := cmd.Flags().GetBool("json")

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

			pName := args[0]

			env := composer.GetEnvByProjectName(pName)
			if env == nil {
				fmt.Println("Project " + pName + " not found")
				os.Exit(1)
			}

			proj := env.GetProjectByName(pName)

			if len(envs) > 0 {
				evars := specs.NewEnvVars()
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

				proj.AddEnvironment(evars)
			}

			compiler, err := template.NewProjectTemplateCompiler(env, proj)
			if err != nil {
				fmt.Println("Error on initialize compiler: " + err.Error())
				os.Exit(1)
			}

			var out string

			if jsonFormat {

				// TODO: Found the issue present on yaml/yamlgo libs
				m := dyno.ConvertMapI2MapS(*compiler.GetVars()).(map[string]interface{})
				y, err := yaml.Marshal(m)
				//y, err := yaml.Marshal(*proj)
				if err != nil {
					fmt.Println("Error on convert vars to yaml: " + err.Error())
					os.Exit(1)
				}
				//data, err := json.Marshal(proj)
				data, err := yamlgo.YAMLToJSON(y)
				if err != nil {
					fmt.Println("Error on convert vars to JSON: " + err.Error())
					os.Exit(1)
				}

				out = string(data)

			} else {
				m := dyno.ConvertMapI2MapS(*compiler.GetVars())
				data, err := yaml.Marshal(m)
				if err != nil {
					fmt.Println("Error on convert vars to yaml: " + err.Error())
					os.Exit(1)
				}

				out = string(data)
			}

			fmt.Println(string(out))
		},
	}

	flags := cmd.Flags()
	flags.Bool("json", false, "Dump variables in JSON format.")
	flags.StringArrayVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	flags.StringArrayVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")

	return cmd
}
