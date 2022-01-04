/*

Copyright (C) 2020-2021 Daniele Rondina <geaaru@sabayonlinux.org>
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
package cmd_command

import (
	"errors"
	"fmt"
	"os"

	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func destroyProject(composer *loader.LxdCInstance, proj string) {
	err := composer.DestroyProject(proj)
	if err != nil {
		fmt.Println("Error on destroy project " + proj + ": " + err.Error())
	}
}

func ApplyCommand(c *specs.LxdCCommand, composer *loader.LxdCInstance,
	proj *specs.LxdCProject, envs []string, varsfiles []string) error {

	err := c.PrepareProject(proj)
	if err != nil {
		return err
	}

	composer.SetFlagsDisabled(c.DisableFlags)
	composer.SetFlagsEnabled(c.EnableFlags)
	composer.SetGroupsDisabled(c.DisableGroups)
	composer.SetGroupsEnabled(c.EnableGroups)
	composer.SetSkipSync(c.SkipSync)
	composer.SetNodesPrefix(c.NodesPrefix)

	if len(varsfiles) > 0 {
		for _, varFile := range varsfiles {
			err := proj.LoadEnvVarsFile(varFile)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on load additional envs var file %s: %s",
						varFile, err.Error(),
					))
			}
		}
	}

	if len(envs) > 0 {
		evars := specs.NewEnvVars()
		for _, e := range envs {
			err := evars.AddKVAggregated(e)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on elaborate var string %s: %s",
						e, err.Error(),
					))
			}
		}

		proj.AddEnvironment(evars)
	}

	if c.GetDestroy() {
		defer destroyProject(composer, proj.GetName())
	}

	err = composer.ApplyProject(proj.GetName())
	if err != nil {
		return errors.New(
			fmt.Sprintf(
				"Error on apply project %s: %s",
				proj.GetName(), err.Error()),
		)
	}

	return nil
}

func NewRunCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string
	var envs []string
	var varsFiles []string
	var commandFiles []string

	var cmd = &cobra.Command{
		Use:     "run <project> <command>",
		Aliases: []string{"r"},
		Short:   "Run a command of environment commands.",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				fmt.Println("Invalid argument. You need <project> and <command>.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			destroy, _ := cmd.Flags().GetBool("destroy")
			prefix, _ := cmd.Flags().GetString("nodes-prefix")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// We need set this before loading phase
			err := config.SetRenderEnvs(renderEnvs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			pname := args[0]
			cname := args[1]

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			env := composer.GetEnvByProjectName(pname)
			if env == nil {
				fmt.Println("No project found with name " + pname)
				os.Exit(1)
			}

			// Load runtime commands
			if len(commandFiles) > 0 {
				for _, f := range commandFiles {
					c, err := specs.CommandFromFile(f)
					if err != nil {
						fmt.Println(fmt.Sprintf("Error on load command file %s: %s",
							f, err.Error()))
						os.Exit(1)
					}

					// Check if the command is already present.
					// NOTE: If this is slow it's better move to a map.
					ce, _ := env.GetCommand(c.Name)
					if ce != nil {
						cmds := []specs.LxdCCommand{}

						for _, cmd := range env.Commands {
							if cmd.Name != c.Name {
								cmds = append(cmds, cmd)
							}
						}
						env.Commands = cmds
					}
					env.AddCommand(c)
				}
			}

			command, err := env.GetCommand(cname)
			if err != nil {
				fmt.Println("No command available with name " + cname +
					" on project " + pname)
				os.Exit(1)
			}

			if destroy {
				command.Destroy = destroy
			}

			if prefix != "" {
				command.NodesPrefix = prefix
			}

			err = ApplyCommand(command, composer,
				env.GetProjectByName(pname),
				envs, varsFiles,
			)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Println("All done.")
		},
	}

	flags := cmd.Flags()
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	flags.StringSliceVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	flags.Bool("destroy", false, "Destroy the selected groups at the end.")
	flags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	flags.StringSliceVar(&varsFiles, "vars-file", []string{},
		"Add additional environments vars file.")
	flags.StringSliceVar(&commandFiles, "command-file", []string{},
		"Add additional commands file.")

	return cmd
}
