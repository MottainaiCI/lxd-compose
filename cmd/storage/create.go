/*

Copyright (C) 2020-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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
		Use:     "create [project] [storage1] [storage2]",
		Short:   "create LXD Storages defined on environment to a specific endpoint or to all groups.",
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
			storages := []specs.LxdCStorage{}

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
				storages = *env.GetStorages()
			} else {
				// Retrieve storage data

				for _, sto := range args[1:] {
					n, err := env.GetStorage(sto)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					storages = append(storages, n)
				}

			}

			if len(storages) == 0 {
				fmt.Println("No storages available.")
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

				for _, sto := range storages {

					isPresent, err := executor.IsPresentStorage(sto.Name)
					if err != nil {
						fmt.Println("Error on check if storage " + sto.Name + " is already present: " +
							err.Error())
						os.Exit(1)
					}

					if !isPresent {
						err := executor.CreateStorage(sto)
						if err != nil {
							fmt.Println("Error on create storage " + sto.Name + ": " + err.Error())
							os.Exit(1)
						}
					} else if upd {
						err := executor.UpdateStorage(sto)
						if err != nil {
							fmt.Println("Error on update storage " + sto.Name + ": " + err.Error())
							os.Exit(1)
						}
					}
				}
			} else {
				// Create storage to all groups

				grpMap := make(map[string]bool, 0)

				for _, grp := range project.Groups {

					if _, ok := grpMap[grp.Connection]; ok {
						// The storage is been created. Nothing to do.
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

					for _, sto := range storages {

						isPresent, err := executor.IsPresentStorage(sto.Name)
						if err != nil {
							fmt.Println("Error on check if storage " + sto.Name + " is already present: " +
								err.Error())
							os.Exit(1)
						}

						if !isPresent {
							err := executor.CreateStorage(sto)
							if err != nil {
								fmt.Println("Error on create storage " + sto.Name + ": " + err.Error())
								os.Exit(1)
							}
							fmt.Println("Storage " + sto.Name + " created.")
						} else {
							if upd {
								err := executor.UpdateStorage(sto)
								if err != nil {
									fmt.Println("Error on update storage " + sto.Name + ": " + err.Error())
									os.Exit(1)
								}
								fmt.Println("Storage " + sto.Name + " updated.")
							} else {
								fmt.Println("Storage " + sto.Name + " already present. Nothing to do.")
							}
						}
					}

				}

			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.BoolP("all", "a", false, "Create all available storages.")
	pflags.BoolP("update", "u", false, "Update the storage if it's already present.")
	pflags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")

	return cmd
}
