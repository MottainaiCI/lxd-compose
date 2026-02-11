/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_security

import (
	"encoding/base64"
	"fmt"
	"os"

	helpers_sec "github.com/MottainaiCI/lxd-compose/pkg/helpers/security"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewDecryptCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "decrypt",
		Aliases: []string{"d", "de"},
		Short:   "Decrypt variables file.",
		PreRun: func(cmd *cobra.Command, args []string) {
			file, _ := cmd.Flags().GetString("vars-file")
			secretsFile, _ := cmd.Flags().GetString("secrets-file")
			if file == "" && secretsFile == "" {
				fmt.Println("Missed mandatory --vars-file or --secrets-file flag")
				os.Exit(1)
			}

			if file != "" && secretsFile != "" {
				fmt.Println("--vars-file and --secrets-file could not be used together")
				os.Exit(1)
			}

			if config.GetSecurity().Key == "" {
				fmt.Println("Encryption key not configured")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			varsfile, _ := cmd.Flags().GetString("vars-file")
			secretsFile, _ := cmd.Flags().GetString("secrets-file")
			to, _ := cmd.Flags().GetString("to")

			keyBytes, err := base64.StdEncoding.DecodeString(config.GetSecurity().Key)
			if err != nil {
				fmt.Println("error on decode key: %s", err.Error())
				os.Exit(1)
			}

			isforVarfile := true
			file := varsfile
			if secretsFile != "" {
				isforVarfile = false
				file = secretsFile
			}

			content, err := os.ReadFile(file)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on read file %s: %s",
					file, err.Error()))
				os.Exit(1)
			}

			var encryptedContent []byte

			if isforVarfile {
				// Create EnvVars object
				evars, err := specs.EnvVarsFromYaml(content)
				if err != nil {
					fmt.Println(fmt.Sprintf("Error on parse file %s: %s",
						file, err.Error()))
					os.Exit(1)
				}

				if !evars.Encrypted {
					fmt.Println(fmt.Sprintf("Var file %s not encrypted.", file))
					os.Exit(1)
				}

				encryptedContent, err = base64.StdEncoding.DecodeString(
					evars.EncryptedContent,
				)

			} else {

				encryptedContent, err = base64.StdEncoding.DecodeString(
					string(content),
				)

			}

			if err != nil {
				fmt.Println(fmt.Sprintf("error on decode base64 encrypted content: %s",
					err.Error()))
				os.Exit(1)
			}

			dkaOpts := helpers_sec.NewDKAOptsDefault()
			if config.GetSecurity().DKAOpts != nil {
				if config.GetSecurity().DKAOpts.TimeIterations != nil {
					dkaOpts.TimeIterations = *config.GetSecurity().DKAOpts.TimeIterations
				}
				if config.GetSecurity().DKAOpts.MemoryUsage != nil {
					dkaOpts.MemoryUsage = *config.GetSecurity().DKAOpts.MemoryUsage
				}
				if config.GetSecurity().DKAOpts.KeyLength != nil {
					dkaOpts.KeyLength = *config.GetSecurity().DKAOpts.KeyLength
				}
				if config.GetSecurity().DKAOpts.Parallelism != nil {
					dkaOpts.Parallelism = *config.GetSecurity().DKAOpts.Parallelism
				}
			}
			decodedBytes, err := helpers_sec.Decrypt(encryptedContent, keyBytes, dkaOpts)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on decrypt content of the file %s: %s",
					file, err.Error()))
				os.Exit(1)
			}

			if to == "" {
				fmt.Println(string(decodedBytes))
			} else {
				err = os.WriteFile(to, decodedBytes, 0644)
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
	pflags.String("secrets-file", "", "Path of the secrets file to encrypt.")
	pflags.String("to", "", "Path of the vars file to generate (stdout if not defined).")

	return cmd
}
