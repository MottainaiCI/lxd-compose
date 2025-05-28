/*
Copyright © 2020-2025 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd_trust

import (
	"fmt"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor"
	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
)

func NewCreateCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string

	var cmd = &cobra.Command{
		Use:     "create [project] [cert1] [cert2]",
		Aliases: []string{"c"},
		Short:   "create LXD certificates available on environment to a specific endpoint or to all groups.",
		PreRun: func(cmd *cobra.Command, args []string) {
			all, _ := cmd.Flags().GetBool("all")
			if len(args) == 0 {
				fmt.Println("Missing project name.")
				os.Exit(1)
			}

			if len(args) > 1 && all {
				fmt.Println("Both profiles and --all option used.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			confdir, _ := cmd.Flags().GetString("lxd-config-dir")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// We need set this before loading phase
			err := config.SetRenderEnvs(renderEnvs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			endpoint, _ := cmd.Flags().GetString("endpoint")
			all, _ := cmd.Flags().GetBool("all")
			//upd, _ := cmd.Flags().GetBool("update")

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			proj := args[0]
			certs := []specs.LxdCCertificate{}

			if confdir == "" {
				confdir = composer.GetConfig().GetGeneral().LxdConfDir
			}

			// Retrieve project
			env := composer.GetEnvByProjectName(proj)
			if env == nil {
				fmt.Println("Project " + proj + " not found")
				os.Exit(1)
			}

			if all {
				certs = *env.GetCertificates()
			} else {
				// Retrieve profiles data

				for _, cname := range args[1:] {
					cert, err := env.GetCertificate(cname)
					if err != nil {
						fmt.Println(err.Error())
						os.Exit(1)
					}

					certs = append(certs, cert)
				}

			}

			if len(certs) == 0 {
				fmt.Println("No certificates available.")
				os.Exit(0)
			}

			if endpoint != "" {

				executor := executor.NewLxdCExecutor(endpoint, confdir, nil, true,
					config.GetLogging().CmdsOutput,
					config.GetLogging().RuntimeCmdsOutput)
				err = executor.Setup()
				if err != nil {
					fmt.Println("Error on setup executor:" + err.Error() + "\n")
					os.Exit(1)
				}

				for _, cert := range certs {
					isPresent, err := executor.IsPresentCertificate(cert.Name)
					if err != nil {
						fmt.Println("Error on check if certificate " + cert.Name + " is already present: " +
							err.Error())
						os.Exit(1)
					}

					if !isPresent {
						err = executor.CreateCertificate(&cert)
						if err != nil {
							fmt.Println("Error on create certificate " + cert.Name + ": " + err.Error())
							os.Exit(1)
						}
						fmt.Println("Certificate " + cert.Name + " created correctly.")
					} else {
						fmt.Println("Certificate " + cert.Name + " already present. Nothing to do.")
					}
				}
			} else {
				// Create profiles to all groups

				for _, proj := range *env.GetProjects() {

					for _, grp := range proj.Groups {

						executor := executor.NewLxdCExecutor(grp.Connection, confdir, nil, true,
							config.GetLogging().CmdsOutput,
							config.GetLogging().RuntimeCmdsOutput)
						err = executor.Setup()
						if err != nil {
							fmt.Println("Error on setup executor for group " + grp.Name + ":" + err.Error() + "\n")
							os.Exit(1)
						}

						for _, cert := range certs {

							isPresent, err := executor.IsPresentCertificate(cert.Name)
							if err != nil {
								fmt.Println("Error on check if certificate " + cert.Name + " is already present: " +
									err.Error())
								os.Exit(1)
							}

							if !isPresent {
								err = executor.CreateCertificate(&cert)
								if err != nil {
									fmt.Println("Error on create certificate " + cert.Name + ": " + err.Error())
									os.Exit(1)
								}
								fmt.Println("Certificate " + cert.Name + " created correctly.")
							} else {
								fmt.Println("Certificate " + cert.Name + " already present. Nothing to do.")
							}
						}

					}

				}

			}

		},
	}

	pflags := cmd.Flags()
	pflags.StringP("endpoint", "e", "", "Set endpoint of the LXD connection")
	pflags.BoolP("all", "a", false, "Create all available certificates.")
	pflags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")

	return cmd
}
