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
package cmd

import (
	"fmt"
	"os"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newPrintCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "print",
		Short: "Show configuration params",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			data, err := config.Yaml()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Println(string(data))
		},
	}

	return cmd
}
