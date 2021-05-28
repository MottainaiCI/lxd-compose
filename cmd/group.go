/*

Copyright (C) 2020-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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
	. "github.com/MottainaiCI/lxd-compose/cmd/group"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newGroupCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "group [command] [OPTIONS]",
		Aliases: []string{"g"},
		Short:   "Execute specific operations for groups",
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		NewListCommand(config),
	)

	return cmd
}
