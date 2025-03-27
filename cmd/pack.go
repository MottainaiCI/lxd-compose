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

	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tarf "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	"github.com/spf13/cobra"
)

func newPackCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var envs []string
	var renderEnvs []string
	var varsFiles []string

	var cmd = &cobra.Command{
		Use:   "pack [project1] ... [projectN] [OPTIONS]",
		Short: "Create a tarball with all needed files for bootstrap projects.",
		PreRun: func(cmd *cobra.Command, args []string) {
			to, _ := cmd.Flags().GetString("to")

			if len(args) == 0 {
				fmt.Println("No environments selected.")
				os.Exit(1)
			}

			if to == "" {
				fmt.Println("Missing mandatory --to option.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			to, _ := cmd.Flags().GetString("to")
			sourceCommonPath, _ := cmd.Flags().GetString("source-common-path")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// We need set this before loading phase
			err := config.SetRenderEnvs(renderEnvs)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			err = composer.LoadEnvironments()
			if err != nil {
				fmt.Println("Error on load environments:" + err.Error() + "\n")
				os.Exit(1)
			}

			// Initialize default tarformers instance
			// to use the config object used by the library.
			cfg := tarf_specs.NewConfig(config.Viper)
			cfg.GetGeneral().Debug = config.GetGeneral().Debug
			cfg.GetLogging().Level = "warning"

			t := tarf.NewTarFormersWithLog(cfg, true)
			tarf.SetDefaultTarFormers(t)

			commonSourceDir, err := composer.PackProjects(
				to, sourceCommonPath, args)
			if err != nil {
				fmt.Println("Error on pack environments: " + err.Error())
				os.Exit(1)
			}

			composer.Logger.InfoC(
				fmt.Sprintf("Tarball %s generated.",
					composer.Logger.Aurora.Bold(to)))
			if commonSourceDir != "" {
				composer.Logger.InfoC(fmt.Sprintf(
					"The source dir to use is: %s",
					composer.Logger.Aurora.Bold(commonSourceDir)))
			}
		},
	}

	flags := cmd.Flags()

	flags.String("to", "", "Path of the tarball to generate.")
	flags.String("source-common-path", "",
		"Define the directory path common for all templates that could be reduce.")
	flags.StringArrayVar(&envs, "env", []string{},
		"Append project environments in the format key=value.")
	flags.StringSliceVar(&varsFiles, "vars-file", []string{},
		"Add additional environments vars file.")
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Append render engine environments in the format key=value.")

	return cmd
}
