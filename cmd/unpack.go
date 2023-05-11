/*
Copyright (C) 2020-2023  Daniele Rondina <geaaru@funtoo.org>
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
	"strings"

	loader "github.com/MottainaiCI/lxd-compose/pkg/loader"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tarf "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	tarf_tools "github.com/geaaru/tar-formers/pkg/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type RenderRaw struct {
	Keys map[string]interface{} `yaml:",inline"`
}

func (r *RenderRaw) Write(file string) error {
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0640)
}

func updateRenderFile(file string, envs *[]string) error {
	renderRaw := &RenderRaw{Keys: make(map[string]interface{}, 0)}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, renderRaw); err != nil {
		return fmt.Errorf("Error on unmarshl file %s: %s",
			file, err.Error())
	}

	for _, e := range *envs {
		if strings.Index(e, "=") < 0 {
			return fmt.Errorf("Invalid KV for render env %s", e)
		}

		key := e[0:strings.Index(e, "=")]
		value := e[strings.Index(e, "=")+1:]

		renderRaw.Keys[key] = value
	}

	return renderRaw.Write(file)
}

func newUnpackCommand(config *specs.LxdComposeConfig) *cobra.Command {
	var renderEnvs []string

	var cmd = &cobra.Command{
		Use:   "unpack [tarball] [OPTIONS]",
		Short: "Unpack a packed lxd-compose tarball.",
		Long: `This command simply unpacks the tarball in input as
an alternative to existing commands like tar.

$> lxd-compose unpack /tmp/myproj.tar.gz --render-file render/default.yaml \
	--render-env "source_base_dir=sources/"
`,
		PreRun: func(cmd *cobra.Command, args []string) {
			renderFile, _ := cmd.Flags().GetString("render-file")
			if len(args) == 0 {
				fmt.Println("No tarball file defined.")
				os.Exit(1)
			}

			if renderFile != "" && len(renderEnvs) == 0 {
				fmt.Println("--render-file option used without --render-env.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			to, _ := cmd.Flags().GetString("to")
			renderFile, _ := cmd.Flags().GetString("render-file")
			sameOwner, _ := cmd.Flags().GetBool("same-owner")

			// Create Instance
			composer := loader.NewLxdCInstance(config)

			// Initialize default tarformers instance
			// to use the config object used by the library.
			cfg := tarf_specs.NewConfig(config.Viper)
			cfg.GetGeneral().Debug = config.GetGeneral().Debug
			cfg.GetLogging().Level = "warning"

			t := tarf.NewTarFormersWithLog(cfg, true)
			tarf.SetDefaultTarFormers(t)

			s := tarf_specs.NewSpecFile()
			s.SameOwner = sameOwner
			opts := tarf_tools.NewTarReaderCompressionOpts(true)
			err := tarf_tools.PrepareTarReader(args[0], opts)
			if err != nil {
				fmt.Println("Error on prepare tar reader:", err.Error())
				os.Exit(1)
			}

			if opts.CompressReader != nil {
				t.SetReader(opts.CompressReader)
			} else {
				t.SetReader(opts.FileReader)
			}

			if to == "" {
				to = "."
			}

			err = t.RunTask(s, to)
			opts.Close()
			if err != nil {
				fmt.Println("Error on untar tarball:", err.Error())
				os.Exit(1)
			}

			if renderFile != "" {
				err = updateRenderFile(renderFile, &renderEnvs)
				if err != nil {
					fmt.Println("Error on update render file:", err.Error())
					os.Exit(1)
				}
				composer.Logger.InfoC(
					fmt.Sprintf("Render file %s updated correctly.",
						composer.Logger.Aurora.Bold(renderFile)))
			}

			composer.Logger.InfoC(":champagne:Operation completed.")
		},
	}

	flags := cmd.Flags()

	flags.String("to", "",
		"Path where unpack the tarball. Default is $PWD.")
	flags.String("render-file", "",
		"Render file where update the source base dir.")
	flags.StringSliceVar(&renderEnvs, "render-env", []string{},
		"Render env variable to replace with the new value.")
	flags.Bool("same-owner", false,
		"Maintain original uid/gid from the tarball.")

	return cmd
}
