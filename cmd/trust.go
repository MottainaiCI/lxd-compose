/*
Copyright © 2020-2025 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd

import (
	. "github.com/MottainaiCI/lxd-compose/cmd/trust"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newTrustCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "trust [command] [OPTIONS]",
		Short:   "Execute specific operations for LXD Certificate",
		Aliases: []string{"tru", "t"},
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		NewListCommand(config),
		NewCreateCommand(config),
	)

	return cmd
}
