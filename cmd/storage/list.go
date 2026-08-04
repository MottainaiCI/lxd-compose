/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

func NewListCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "list <project>",
		Aliases: []string{"l"},
		Short:   "List the definitions of the available storages defined in the project.",

		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No project selected.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			project := args[0]
			env := composer.GetEnvByProjectName(project)
			if env == nil {
				fmt.Println("Project not found")
				os.Exit(1)
			}

			storages := *env.GetStorages()

			if search != "" {
				nstorages := []specs.LxdCStorage{}

				for _, s := range storages {
					res := helpers.RegexEntry(search, []string{s.GetName()})
					if len(res) > 0 {
						nstorages = append(nstorages, s)
					}
				}

				storages = nstorages
			}

			if jsonOutput {

				data, _ := json.Marshal(storages)
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
					tablewriter.WithRowAutoWrap(tw.WrapNone),
				)
				table.Header([]string{
					"Storages", "Driver", "Documentation",
				})

				for _, s := range storages {
					table.Append([]string{
						s.GetName(),
						s.GetDriver(),
						s.GetDocumentation(),
					})
				}
				table.Render()
			}
		},
	}

	pflags := cmd.Flags()
	pflags.Bool("json", false, "JSON output")
	pflags.StringP("search", "s", "", "Regex filter to use with storage name.")

	return cmd
}
