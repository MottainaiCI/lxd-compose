/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd

import (
	. "github.com/MottainaiCI/lxd-compose/cmd/security"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func newSecurityCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "security [command] [OPTIONS]",
		Aliases: []string{"se", "sec"},
		Short:   "Execute security operations.",
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		NewEncryptCommand(config),
		NewDecryptCommand(config),
		NewGenKeyCommand(config),
	)

	return cmd
}
