/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
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

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2020 Mottainai - Daniele Rondina

Mottainai - LXD Compose Integrator`

	LXD_COMPOSE_VERSION = `0.1.0`
)

func initConfig(config *specs.LxdComposeConfig) {
	// Set env variable
	config.Viper.SetEnvPrefix(specs.LXD_COMPOSE_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(specs.LXD_COMPOSE_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)

}

func initCommand(rootCmd *cobra.Command, config *specs.LxdComposeConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "LXD Compose configuration file")
	pflags.String("lxd-config-dir", "", "Override LXD config directory.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("general.lxd_confdir", pflags.Lookup("lxd-config-dir"))

	rootCmd.AddCommand(
		newApplyCommand(config),
		newDestroyCommand(config),
		newPrintCommand(config),
		newValidateCommand(config),
		newCompileCommand(config),
		newNodeCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.LxdComposeConfig = specs.NewLxdComposeConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      LXD_COMPOSE_VERSION,
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			v.SetConfigType("yml")
			if v.Get("config") == "" {
				config.Viper.AddConfigPath(".")
			} else {
				v.SetConfigFile(v.Get("config").(string))
			}

			// Parse configuration file
			err = config.Unmarshal()
			utils.CheckError(err)
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
