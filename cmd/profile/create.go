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
package cmd_profile

import (
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewCreateCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "create [project] [profile1] [profile2]",
		Aliases: []string{"c"},
		Short:   "create LXD profiles available on environment to a specific endpoint or to all groups.",
		PreRun: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			if len(args) == 0 {
				fmt.Println("Missing project name.")
				os.Exit(1)
			}

			if len(args) > 1 && all {
				fmt.Println("Both profiles and --all option used.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")

			// Create Instance
			composer := loader.NewLxdCInstance(config)
			endpoint, _ := cmd.Flags().GetString("endpoint")
			all, _ := cmd.Flags().GetBool("all")
			upd, _ := cmd.Flags().GetBool("update")

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			proj := args[0]
			profiles := []specs.LxdCProfile{}

			if confdir == "" {
				confdir = composer.GetConfig().GetGeneral().LxdConfDir
			}

			// Retrieve project
			env := composer.GetEnvByProjectName(proj)
			if env == nil {
				fmt.Println("Project " + proj + " not found")
				os.Exit(1)
			}

			if all {
				profiles = *env.GetProfiles()
			} else {
				// Retrieve profiles data

				for _, prof := range args[1:] {
					p, err := env.GetProfile(prof)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					profiles = append(profiles, p)
				}

			}

			if len(profiles) == 0 {
				fmt.Println("No profiles available.")
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

				for _, prof := range profiles {
					isPresent, err := executor.IsPresentProfile(prof.Name)
					if err != nil {
						fmt.Println("Error on check if profile " + prof.Name + " is already present: " +
							err.Error())
						os.Exit(1)
					}

					if !isPresent {
						err := executor.CreateProfile(prof)
						if err != nil {
							fmt.Println("Error on create profile " + prof.Name + ": " + err.Error())
							os.Exit(1)
						}
						fmt.Println("Profile " + prof.Name + " created correctly.")
					} else if upd {
						err := executor.UpdateProfile(prof)
						if err != nil {
							fmt.Println("Error on update profile " + prof.Name + ": " + err.Error())
							os.Exit(1)
						}
						fmt.Println("Profile " + prof.Name + " updated correctly.")
					} else {
						fmt.Println("Profile " + prof.Name + " already present. Nothing to do.")
					}
				}
			} else {
				// Create profiles to all groups

				for _, proj := range *env.GetProjects() {

					for _, grp := range proj.Groups {

						executor := executor.NewLxdCExecutor(grp.Connection, confdir, nil, true,
							config.GetLogging().CmdsOutput,
							config.GetLogging().RuntimeCmdsOutput)
						err = executor.Setup()
						if err != nil {
							fmt.Println("Error on setup executor for group " + grp.Name + ":" + err.Error() + "\n")
							os.Exit(1)
						}

						for _, prof := range profiles {

							isPresent, err := executor.IsPresentProfile(prof.Name)
							if err != nil {
								fmt.Println("Error on check if profile " + prof.Name + " is already present: " +
									err.Error())
								os.Exit(1)
							}

							if !isPresent {
								err := executor.CreateProfile(prof)
								if err != nil {
									fmt.Println("Error on create profile " + prof.Name + ": " + err.Error())
									os.Exit(1)
								}
								fmt.Println("Profile " + prof.Name + " created correctly.")
							} else if upd {
								err := executor.UpdateProfile(prof)
								if err != nil {
									fmt.Println("Error on update profile " + prof.Name + ": " + err.Error())
									os.Exit(1)
								}
								fmt.Println("Profile " + prof.Name + " updated correctly.")
							} else {
								fmt.Println("Profile " + prof.Name + " already present. Nothing to do.")
							}
						}

					}

				}

			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.BoolP("all", "a", false, "Create all available profiles.")
	pflags.BoolP("update", "u", false, "Update the profiles if it's already present.")

	return cmd
}
