/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_security

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
)

func NewEncryptCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"e", "enc"},
		Short:   "Encrypt variables file.",
		PreRun: func(cmd *cobra.Command, args []string) {
			file, _ := cmd.Flags().GetString("vars-file")
			if file == "" {
				fmt.Println("Missed mandatory --vars-file flag")
				os.Exit(1)
			}

			if config.GetSecurity().Key == "" {
				fmt.Println("Encryption key not configured")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			file, _ := cmd.Flags().GetString("vars-file")
			to, _ := cmd.Flags().GetString("to")

			keyBytes, err := base64.StdEncoding.DecodeString(config.GetSecurity().Key)
			if err != nil {
				fmt.Println("error on decode key: %s", err.Error())
				os.Exit(1)
			}

			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on read file %s: %s",
					file, err.Error()))
				os.Exit(1)
			}

			dkaOpts := helpers.NewDKAOptsDefault()
			encryptedFile, err := helpers.Encrypt(content, keyBytes, dkaOpts)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on encrypt content of the file %s: %s",
					file, err.Error()))
				os.Exit(1)
			}

			evars := specs.NewEnvVars()
			evars.Encrypted = true
			evars.EncryptedContent = base64.StdEncoding.EncodeToString(encryptedFile)

			data, err := yaml.Marshal(evars)
			if err != nil {
				fmt.Println("Error on marshalling generated EnvVars: ", err.Error())
				os.Exit(1)
			}

			if to == "" {
				fmt.Println(string(data))
			} else {
				err = os.WriteFile(to, data, 0644)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on write file %s: %s",
						to, err.Error()))
					os.Exit(1)
				}
			}
		},
	}

	pflags := cmd.Flags()
	pflags.String("vars-file", "", "Path of the vars file to encrypt.")
	pflags.String("to", "", "Path of the vars file to generate (stdout if not defined).")

	return cmd
}
