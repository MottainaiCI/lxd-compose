/*
Copyright © 2020-2025 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_diagnose

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
		Use:     "list",
		Aliases: []string{"l", "li"},
		Short:   "List all loaded projects.",
		Run: func(cmd *cobra.Command, args []string) {
			projects := []specs.LxdCProjectSanitized{}

			jsonOutput, _ := cmd.Flags().GetBool("json")
			search, _ := cmd.Flags().GetString("search")

			// Create Instance
			composer := loader.NewLxdCInstance(config)
			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			for _, e := range *composer.GetEnvironments() {
				for _, p := range *e.GetProjects() {
					if search != "" {
						res := helpers.RegexEntry(search, []string{p.GetName()})
						if len(res) > 0 {
							projects = append(projects, *p.Sanitize())
						}
					} else {
						projects = append(projects, *p.Sanitize())
					}
				}
			}

			if jsonOutput {

				data, err := json.Marshal(projects)
				if err != nil {
					fmt.Println("Error on decode projects ", err.Error())
					os.Exit(1)
				}
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
					tablewriter.WithRowAutoWrap(tw.WrapNone),
				)
				table.Header([]string{
					"Project Name", "Description", "# Groups",
				})

				for _, p := range projects {

					table.Append([]string{
						p.GetName(),
						p.GetDescription(),
						fmt.Sprintf("%d", len(*p.GetGroups())),
					})
				}

				table.Render()
			}

		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")
	flags.StringP("search", "s", "", "Regex filter to use with network name.")

	return cmd
}
