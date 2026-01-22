/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewGenKeyCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "genkey",
		Aliases: []string{"g", "gk"},
		Short:   "Generate an encryption key.",
		PreRun: func(cmd *cobra.Command, args []string) {
			lenKey, _ := cmd.Flags().GetUint64("length")
			if lenKey < 32 {
				fmt.Println("length of the key to small. Minimal 32 bytes.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			lenKey, _ := cmd.Flags().GetUint64("length")

			key := make([]byte, lenKey)
			_, err := rand.Read(key)
			if err != nil {
				fmt.Println("Error on generate key: " + err.Error())
				os.Exit(1)
			}

			fmt.Println(base64.StdEncoding.EncodeToString(key))
		},
	}

	pflags := cmd.Flags()
	pflags.Uint64P("length", "l", 64, "Define the length of the key")

	return cmd
}
