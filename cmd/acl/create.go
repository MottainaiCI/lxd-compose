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
package cmd_acl

import (
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewCreateCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string

	var cmd = &cobra.Command{
		Use:     "create [project] [acl1] [acl2]",
		Short:   "create LXD ACL defined on environment to a specific endpoint or to all groups.",
		Aliases: []string{"c"},
		PreRun: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			if len(args) == 0 {
				fmt.Println("Missing project name.")
				os.Exit(1)
			}

			if len(args) > 1 && all {
				fmt.Println("Both storages and --all option used.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// We need set this before loading phase
			err := config.SetRenderEnvs(renderEnvs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			endpoint, _ := cmd.Flags().GetString("endpoint")
			all, _ := cmd.Flags().GetBool("all")
			upd, _ := cmd.Flags().GetBool("update")

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			proj := args[0]
			acls := []specs.LxdCAcl{}

			if confdir == "" {
				confdir = composer.GetConfig().GetGeneral().LxdConfDir
			}

			// Retrieve project
			env := composer.GetEnvByProjectName(proj)
			if env == nil {
				fmt.Println("Project " + proj + " not found")
				os.Exit(1)
			}

			project := env.GetProjectByName(proj)

			if all {
				acls = *env.GetACLs()
			} else {
				// Retrieve storage data

				for _, acl := range args[1:] {
					a, err := env.GetACL(acl)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					acls = append(acls, a)
				}

			}

			if len(acls) == 0 {
				fmt.Println("No acls available.")
				os.Exit(0)
			}

			if endpoint != "" {

				executor := executor.NewLxdCExecutor(endpoint, confdir, nil, true,
					config.GetLogging().CmdsOutput,
					config.GetLogging().RuntimeCmdsOutput)
				err = executor.Setup()
				if err != nil {
					fmt.Println("Error on setup executor:" + err.Error() + "\n")
					os.Exit(1)
				}

				for _, acl := range acls {

					isPresent, err := executor.IsPresentACL(acl.Name)
					if err != nil {
						fmt.Println("Error on check if acl " + acl.Name + " is already present: " +
							err.Error())
						os.Exit(1)
					}

					if !isPresent {
						err := executor.CreateACL(&acl)
						if err != nil {
							fmt.Println("Error on create acl " + acl.Name + ": " + err.Error())
							os.Exit(1)
						}
					} else if upd {
						err := executor.UpdateACL(&acl)
						if err != nil {
							fmt.Println("Error on update acl " + acl.Name + ": " + err.Error())
							os.Exit(1)
						}
					}
				}
			} else {
				// Create acl to all groups

				grpMap := make(map[string]bool, 0)

				for _, grp := range project.Groups {

					if _, ok := grpMap[grp.Connection]; ok {
						// The acl is been created. Nothing to do.
						continue
					} else {
						grpMap[grp.Connection] = true
					}

					executor := executor.NewLxdCExecutor(grp.Connection, confdir, nil, true,
						config.GetLogging().CmdsOutput,
						config.GetLogging().RuntimeCmdsOutput)
					err = executor.Setup()
					if err != nil {
						fmt.Println("Error on setup executor for group " + grp.Name + ":" + err.Error() + "\n")
						os.Exit(1)
					}

					for _, acl := range acls {

						isPresent, err := executor.IsPresentACL(acl.Name)
						if err != nil {
							fmt.Println("Error on check if acl " + acl.Name + " is already present: " +
								err.Error())
							os.Exit(1)
						}

						if !isPresent {
							err := executor.CreateACL(&acl)
							if err != nil {
								fmt.Println("Error on create acl " + acl.Name + ": " + err.Error())
								os.Exit(1)
							}
							fmt.Println("ACL " + acl.Name + " created.")
						} else {
							if upd {
								err := executor.UpdateACL(&acl)
								if err != nil {
									fmt.Println("Error on update acl " + acl.Name + ": " + err.Error())
									os.Exit(1)
								}
								fmt.Println("ACL " + acl.Name + " updated.")
							} else {
								fmt.Println("ACL " + acl.Name + " already present. Nothing to do.")
							}
						}
					}

				}

			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.BoolP("all", "a", false, "Create all available acls.")
	pflags.BoolP("update", "u", false, "Update the acl if it's already present.")
	pflags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")

	return cmd
}
