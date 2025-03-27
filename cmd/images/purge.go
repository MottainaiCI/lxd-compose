/*
Copyright (C) 2020-2025  Daniele Rondina <geaaru@macaronios.org>
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
package cmd_images

import (
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewPurgeCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string

	var cmd = &cobra.Command{
		Use:     "purge [project1] ... [projectN]",
		Short:   "Purge LXD Images from one or more groups.",
		Aliases: []string{"p"},
		PreRun: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			if len(args) == 0 && !all && endpoint == "" {
				fmt.Println("Missing project name.")
				os.Exit(1)
			}

			if len(args) > 1 && all || len(args) > 1 && endpoint != "" || all && endpoint != "" {
				fmt.Println("Both projects and --all or --endpoint option used.")
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
			allImages, _ := cmd.Flags().GetBool("all-images")
			withoutAliases, _ := cmd.Flags().GetBool("without-aliases")
			matches, _ := cmd.Flags().GetStringArray("match")
			fprint, _ := cmd.Flags().GetString("fingerprint")

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			if confdir == "" {
				confdir = composer.GetConfig().GetGeneral().LxdConfDir
			}

			remoteMap := make(map[string]bool, 0)

			purgeOpts := &executor.PurgeOpts{
				All:         allImages,
				Fingerprint: fprint,
				Matches:     matches,
				NoAliases:   withoutAliases,
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

				err = executor.PurgeImages(purgeOpts)
				if err != nil {
					fmt.Println("Error on purge images for endpoint " + endpoint + ":" +
						err.Error() + "\n")
					os.Exit(1)
				}
			} else {
				if all {

					for _, env := range *composer.GetEnvironments() {
						for _, proj := range *env.GetProjects() {
							for _, grp := range proj.Groups {
								if _, ok := remoteMap[grp.Connection]; ok {
									// Remote already processed
									continue
								}

								remoteMap[grp.Connection] = true

								executor := executor.NewLxdCExecutor(grp.Connection, confdir, nil, true,
									config.GetLogging().CmdsOutput,
									config.GetLogging().RuntimeCmdsOutput)
								err = executor.Setup()
								if err != nil {
									fmt.Println("Error on setup executor for group " + grp.Name + ":" +
										err.Error() + "\n")
									os.Exit(1)
								}

								err = executor.PurgeImages(purgeOpts)
								if err != nil {
									fmt.Println("Error on purge images for group " + grp.Name + ":" +
										err.Error() + "\n")
									os.Exit(1)
								}
							}

						}
					}
				} else {

					for _, pstring := range args {

						// Retrieve project
						env := composer.GetEnvByProjectName(pstring)
						if env == nil {
							fmt.Println("Project " + pstring + " not found")
							os.Exit(1)
						}

						proj := env.GetProjectByName(pstring)

						for _, grp := range proj.Groups {
							if _, ok := remoteMap[grp.Connection]; ok {
								// Remote already processed
								continue
							}

							remoteMap[grp.Connection] = true

							executor := executor.NewLxdCExecutor(grp.Connection, confdir, nil, true,
								config.GetLogging().CmdsOutput,
								config.GetLogging().RuntimeCmdsOutput)
							err = executor.Setup()
							if err != nil {
								fmt.Println("Error on setup executor for group " + grp.Name + ":" +
									err.Error() + "\n")
								os.Exit(1)
							}

							err = executor.PurgeImages(purgeOpts)
							if err != nil {
								fmt.Println("Error on purge images for group " + grp.Name + ":" +
									err.Error() + "\n")
								os.Exit(1)
							}
						}

					}

				}
			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.BoolP("all", "a", false, "Purge images from all projects.")
	pflags.Bool("all-images", false, "Purge all images.")
	pflags.Bool("without-aliases", false, "Purge all images without aliases.")
	pflags.StringP("fingerprint", "f", "",
		"Delete image of specified fingerprint")
	pflags.StringArray("match", []string{},
		"Define one or more regex for select images to purge by aliases.")
	pflags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")

	return cmd
}
