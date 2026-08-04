/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "list of LXD networks available to endpoint.",
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			connType, _ := cmd.Flags().GetString("connection-type")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			if confdir == "" {
				confdir = config.GetGeneral().LxdConfDir
			}

			executor := executor.NewLxdCExecutor(
				connType, endpoint, confdir, nil, true,
				config.GetLogging().CmdsOutput,
				config.GetLogging().RuntimeCmdsOutput)
			err = executor.Setup()
			if err != nil {
				fmt.Println("Error on setup executor:" + err.Error() + "\n")
				os.Exit(1)
			}

			list, err := executor.GetNetworkList()
			if err != nil {
				fmt.Println("Error on retrieve network list: " + err.Error() + "\n")
				os.Exit(1)
			}

			if search != "" {
				list = helpers.RegexEntry(search, list)
			}

			if jsonOutput {
				data, _ := json.Marshal(list)
				fmt.Println(string(data))

			} else {

				table := tablewriter.NewWriter(os.Stdout,
					tablewriter.WithRendition(tw.Rendition{
						Borders: tw.Border{
							Left:   tw.On,
							Top:    tw.Off,
							Right:  tw.On,
							Bottom: tw.Off,
						},
						Symbols: tw.Symbols{
							Merge: "|",
						},
					}),
				)
				table.Header([]string{
					"Network Name", "Description", "Environments Mapped",
				})

				for _, c := range list {

					mappedEnvs := ""
					networkDescr := ""
					envs := composer.GetEnvsUsingNetwork(c)
					if len(envs) > 0 {
						for idx, e := range envs {
							net, _ := e.GetNetwork(c)
							networkDescr = net.Description
							if idx > 0 {
								mappedEnvs += "\n" + filepath.Base(e.File)
							} else {
								mappedEnvs += filepath.Base(e.File)
							}
						}
					}

					table.Append([]string{
						c,
						networkDescr,
						mappedEnvs,
					})

				}

				table.Render()
			}
		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.String("connection-type", "incus", "Set connection type.")
	pflags.Bool("json", false, "JSON output")
	pflags.StringP("search", "s", "", "Regex filter to use with network name.")

	return cmd
}
