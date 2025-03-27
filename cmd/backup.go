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
	"sort"
	"strconv"
	"strings"
	"time"

	lxd_executor "github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newBackupCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var enabledGroups []string
	var disabledGroups []string
	var renderEnvs []string
	var profiles2remove []string

	var cmd = &cobra.Command{
		Use:     "backup [list-of-projects]",
		Short:   "Backup the container of the listed projects.",
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
			maxBackups, _ := cmd.Flags().GetInt("max-backups")
			ret := 0

			composer.SetGroupsDisabled(disabledGroups)
			composer.SetGroupsEnabled(enabledGroups)
			composer.SetNodesPrefix(prefix)

			projects := args[0:]
			mapExecutors := make(map[string]*lxd_executor.LxdCExecutor, 0)

			t := time.Now()
			containerPostfix := t.Format("20060102")

			for _, proj := range projects {

				env := composer.GetEnvByProjectName(proj)
				if env == nil {
					fmt.Println("Project " + proj + " not found")
					os.Exit(1)
				}

				composer.Logger.Info(":rocket:Backup containers for project " + proj + "...")
				pObj := env.GetProjectByName(proj)

				for _, grp := range *pObj.GetGroups() {
					if !grp.ToProcess(enabledGroups, disabledGroups) {
						composer.Logger.Debug("Skipped group ", grp.Name)
						continue
					}

					for _, node := range grp.Nodes {

						key := fmt.Sprintf(
							"%s", grp.Connection,
						)

						executor, ok := mapExecutors[key]

						if !ok {
							// Initialize executor
							executor = lxd_executor.NewLxdCExecutor(grp.Connection,
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
								continue
							}

							executor.SetP2PMode(config.GetGeneral().P2PMode)
							mapExecutors[key] = executor
						}

						backupName := node.GetName() + "-" + containerPostfix

						originPresent, err := executor.IsPresentContainer(node.GetName())
						if err != nil {
							fmt.Println(
								fmt.Sprintf(
									"Error on check if the container %s is present: %s. Skipping.",
									node.GetName(), err.Error()))
							ret = 1
							continue
						}

						if !originPresent {
							fmt.Println(
								fmt.Sprintf(
									"Container %s is not present on group %s. Skipped.",
									node.GetName(), grp.Name,
								))
							continue
						}

						present, err := executor.IsPresentContainer(backupName)
						if err != nil {
							fmt.Println(
								fmt.Sprintf(
									"Error on check if the container %s is present: %s. Skipping.",
									backupName, err.Error()))
							ret = 1
							continue
						}

						if present {
							composer.Logger.InfoC(
								fmt.Sprintf(
									":icecream:%s Container already present :check_mark:.",
									composer.Logger.Aurora.BrightCyan(
										fmt.Sprintf("[%s]", backupName))))
						} else {
							err := executor.CopyContainerOnInstance(
								node.GetName(), backupName,
							)
							if err != nil {
								fmt.Println(
									fmt.Sprintf(
										"Error on check copy container %s: %s. Skipping.",
										node.GetName(), err.Error()))
								ret = 1
								continue
							}

							composer.Logger.InfoC(
								fmt.Sprintf(
									":icecream:%s Container %s copied. :check_mark:",
									composer.Logger.Aurora.Bold(
										composer.Logger.Aurora.BrightCyan(
											fmt.Sprintf("[%s]", backupName))),
									composer.Logger.Aurora.Bold(node.GetName())))

							if len(profiles2remove) > 0 {
								composer.Logger.InfoC(
									fmt.Sprintf(
										":knife:%s Removed profiles: %s.",
										composer.Logger.Aurora.BrightCyan(
											fmt.Sprintf("[%s]", backupName)),
										strings.Join(profiles2remove, ","),
									))

								err = executor.RemoveProfilesFromContainer(
									backupName, profiles2remove,
								)
								if err != nil {
									composer.Logger.Error(
										fmt.Sprintf(
											"Error on removing profiles from %s container: %s",
											backupName, err.Error()),
									)
									ret = 1
									continue
								}
							}

						}

						if maxBackups > 0 {
							// POST: We need to maintain limited number of backups.

							// Retrieve the list of container.
							list, err := executor.GetContainerList()
							if err != nil {
								composer.Logger.Error(
									fmt.Sprintf(
										"Error on retrieve the list of container"))
								ret = 1
								continue
							}

							matchedDates := []string{}

							for _, n := range list {
								if strings.HasPrefix(n, node.GetName()+"-") {
									if len(n) == len(node.GetName())+9 {
										date := n[len(node.GetName())+1:]
										if _, err := strconv.Atoi(date); err == nil {
											matchedDates = append(matchedDates, date)
										}
									}
								}
							}

							if len(matchedDates) > 0 && len(matchedDates) > maxBackups {
								sort.Strings(matchedDates)

								backup2remove := len(matchedDates) - maxBackups
								for i := 0; i < backup2remove; i++ {
									cname := node.GetName() + "-" + matchedDates[i]
									composer.Logger.InfoC(
										composer.Logger.Aurora.Bold(
											fmt.Sprintf(":knife:[%s] Removing container...",
												cname)))

									err := executor.DeleteContainer(cname)
									if err != nil {
										composer.Logger.Warning(
											fmt.Sprintf("error on removing container %s: %s",
												cname, err.Error()))
										ret = 1
									}

								}
							}

						}

					}
				}
			}

			if ret != 0 {
				fmt.Println("Not all containers are been copy correctly.")

			} else {
				composer.Logger.InfoC(
					fmt.Sprintf(":chequered_flag:%s :chequered_flag:",
						composer.Logger.Aurora.Bold("All done!")),
				)
			}
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVar(&disabledGroups, "disable-group", []string{},
		"Skip selected group from deploy.")
	flags.StringSliceVar(&enabledGroups, "enable-group", []string{},
		"Apply only selected groups.")
	flags.StringSliceVar(&profiles2remove, "remove-profile", []string{},
		"Define the profile to remove on the backuped containers.")
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")
	flags.String("nodes-prefix", "", "Customize project nodes name with a prefix")
	flags.Int("max-backups", 0, "Number of max backups to keep. 0 means no limit.")

	return cmd
}
