/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_command

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
	var commandFiles []string

	var cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "li"},
		Short:   "list of environment commands.",
		Run: func(cmd *cobra.Command, args []string) {

			jsonOutput, _ := cmd.Flags().GetBool("json")
			search, _ := cmd.Flags().GetString("search")
			commands := []specs.LxdCCommand{}

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			err := composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			if len(commandFiles) > 0 {
				for _, f := range commandFiles {
					c, err := specs.CommandFromFile(f)
					if err != nil {
						fmt.Println(fmt.Sprintf("Error on load command file %s: %s",
							f, err.Error()))
						os.Exit(1)
					}

					env := composer.GetEnvByProjectName(c.GetProject())
					if env == nil {
						fmt.Println("No project found with name " + c.GetProject() +
							" for add command from cli " + c.GetName())
						os.Exit(1)
					}

					env.AddCommand(c)
				}
			}

			for _, env := range *composer.GetEnvironments() {

				if len(*env.GetCommands()) > 0 {
					if search != "" {
						for _, c := range *env.GetCommands() {
							res := helpers.RegexEntry(search, []string{c.GetName()})
							if len(res) > 0 {
								commands = append(commands, c)
							}
						}
					} else {
						commands = append(commands, *env.GetCommands()...)
					}
				}
			}

			if jsonOutput {

				data, _ := json.Marshal(commands)
				fmt.Println(string(data))

			} else {
				if len(commands) > 0 {
					table := tablewriter.NewWriter(os.Stdout,
						tablewriter.WithRendition(tw.Rendition{
							Borders: tw.Border{
								Left:   tw.On,
								Top:    tw.On,
								Right:  tw.On,
								Bottom: tw.On,
							},
							Symbols: tw.Symbols{
								Merge: "|",
							},
						}),
						tablewriter.WithRowAutoWrap(tw.WrapNone),
						tablewriter.WithColumnMaxWidth(150),
					)
					table.Header([]string{"Command", "Project", "Description"})

					for _, c := range commands {
						table.Append([]string{
							c.Name,
							c.Project,
							c.Description,
						})
					}

					table.Render()

				} else {
					fmt.Println("No commands available.")
				}
			}

		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")
	flags.StringP("search", "s", "", "Regex filter to use with command name.")
	flags.StringSliceVar(&commandFiles, "command-file", []string{},
		"Add additional commands file.")

	return cmd
}
