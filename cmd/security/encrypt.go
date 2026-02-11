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
			encryptedFile, err := helpers_sec.Encrypt(content, keyBytes, dkaOpts)
			if err != nil {
				fmt.Println(fmt.Sprintf("Error on encrypt content of the file %s: %s",
					file, err.Error()))
				os.Exit(1)
			}

			var data []byte
			if isforVarfile {
				evars := specs.NewEnvVars()
				evars.Encrypted = true
				evars.EncryptedContent = base64.StdEncoding.EncodeToString(encryptedFile)

				data, err = yaml.Marshal(evars)
				if err != nil {
					fmt.Println("Error on marshalling generated EnvVars: ", err.Error())
					os.Exit(1)
				}

			} else {
				data = []byte(base64.StdEncoding.EncodeToString(encryptedFile))
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
	pflags.String("secrets-file", "", "Path of the secrets file to encrypt.")
	pflags.String("to", "", "Path of the vars file to generate (stdout if not defined).")

	return cmd
}
