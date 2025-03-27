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
GNU General Public License for more details.:s

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newFetchCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var enabledGroups []string
	var disabledGroups []string
	var renderEnvs []string
	var testProfiles []string

	var cmd = &cobra.Command{
		Use:     "fetch [list-of-projects]",
		Short:   "Fetch images of the nodes defined on specified groups.",
		Aliases: []string{"f"},
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
			testImages, _ := cmd.Flags().GetBool("test-images")
			sleep, _ := cmd.Flags().GetUint("sleep")
			ret := 0

			composer.SetGroupsDisabled(disabledGroups)
			composer.SetGroupsEnabled(enabledGroups)
			composer.SetNodesPrefix(prefix)

			projects := args[0:]
			mapExecutors := make(map[string]*lxd_executor.LxdCExecutor, 0)

			for _, proj := range projects {

				env := composer.GetEnvByProjectName(proj)
				if env == nil {
					fmt.Println("Project " + proj + " not found")
					os.Exit(1)
				}

				composer.Logger.Info("Fetching images for project " + proj + "...")
				pObj := env.GetProjectByName(proj)

				for _, grp := range *pObj.GetGroups() {
					if !grp.ToProcess(enabledGroups, disabledGroups) {
						composer.Logger.Debug("Skipped group ", grp.Name)
						continue
					}

					for _, node := range grp.Nodes {

						key := fmt.Sprintf(
							"%s|%s|%s",
							grp.Connection, node.ImageSource, node.ImageRemoteServer,
						)

						if _, ok := mapExecutors[key]; !ok {

							// Initialize executor
							executor := lxd_executor.NewLxdCExecutor(grp.Connection,
								config.GetGeneral().LxdConfDir, []string{}, grp.Ephemeral,
								config.GetLogging().CmdsOutput,
								config.GetLogging().RuntimeCmdsOutput)
							err := executor.Setup()
							if err != nil {
								fmt.Println(
									fmt.Sprintf(
										"Error on initialize executor for group %s and connection %s: %s",
										grp.GetName(), grp.Connection, err.Error()))
								fmt.Println("Skipping group", grp.Name)
								ret = 1
							}

							executor.SetP2PMode(config.GetGeneral().P2PMode)
							mapExecutors[key] = executor

						}
					}
				}
			}

			if len(mapExecutors) > 0 {

				for key, executor := range mapExecutors {

					// Split key to retrieve needed informations
					imageData := strings.Split(key, "|")
					_, err := executor.PullImage(imageData[1], imageData[2])
					if err != nil {
						composer.Logger.Error(
							fmt.Sprintf("Error on fetch image %s from server %s.",
								imageData[1], imageData[0]))
						ret += 1
					} else if testImages {

						composer.Logger.Info(fmt.Sprintf(
							"[%s] Testing image %s fetched from server %s...",
							imageData[0], imageData[1], imageData[2]))

						err = executor.CreateContainer("test-image",
							imageData[1], imageData[0], testProfiles)
						if err != nil {
							composer.Logger.Error(
								fmt.Sprintf("Error on create container with image %s for group %s: %s",
									imageData[1], key, err.Error()))
						}

						composer.Logger.Info(fmt.Sprintf(
							"[%s] Sleeping for %d seconds...",
							imageData[0], sleep,
						))
						duration, _ := time.ParseDuration(fmt.Sprintf("%ds", sleep))
						time.Sleep(duration)

						err = executor.DeleteContainer("test-image")
						if err != nil {
							composer.Logger.Error(
								fmt.Sprintf("Error on delete container with image %s for group %s: %s",
									imageData[1], imageData[0], err.Error()))
						}

					}

				}
			}

			if ret != 0 {
				fmt.Println("Not all images are been fetched correctly.")

			} else {
				fmt.Println("All done.")
			}

			os.Exit(ret)
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVar(&disabledGroups, "disable-group", []string{},
		"Skip selected group from deploy.")
	flags.StringSliceVar(&enabledGroups, "enable-group", []string{},
		"Apply only selected groups.")
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	flags.String("nodes-prefix", "", "Customize project nodes name with a prefix")

	flags.StringSliceVar(&testProfiles, "test-profile", []string{},
		"Define the list of LXD profile to use on testing container. Used with --test-images")
	flags.Bool("test-images", false, "Testing fetched images.")
	flags.Uint("sleep", 3,
		"Number of seconds sleep before delete the testing container. Used with --test-images.")

	return cmd
}
