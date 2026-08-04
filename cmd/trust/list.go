/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_trust

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "li"},
		Short:   "list of LXD certificates available to endpoint.",
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			connType, _ := cmd.Flags().GetString("connection-type")

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

			list, err := executor.GetCertificates()
			if err != nil {
				fmt.Println("Error on retrieve certificates list: " + err.Error() + "\n")
				os.Exit(1)
			}

			if jsonOutput {
				data, _ := json.Marshal(list)
				fmt.Println(string(data))

			} else {

				table := tablewriter.NewTable(os.Stdout,
					tablewriter.WithRendition(tw.Rendition{
						Borders: tw.Border{
							Left:   tw.On,
							Top:    tw.Off,
							Right:  tw.On,
							Bottom: tw.Off,
						},
						Symbols: tw.NewSymbols(tw.StyleASCII),
					}),
				)
				table.Header([]string{
					"Certificate Name", "Type", "Fingerprint",
				})

				for _, c := range list {

					table.Append([]string{
						c.Name,
						c.Type,
						c.Fingerprint,
					})

				}

				table.Render()
			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the connection.")
	pflags.String("connection-type", "incus", "Override connection type.")
	pflags.Bool("json", false, "JSON output")

	return cmd
}
