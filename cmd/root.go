/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	cliName = `Copyright (c) 2020-2024 Mottainai - Daniele Rondina

Mottainai - LXD Compose Integrator`

	LXD_COMPOSE_VERSION = `0.34.0`
)

var (
	BuildTime      string
	BuildCommit    string
	BuildGoVersion string
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

func cmdNeedConfig(cmd string) bool {
	if cmd != "unpack" {
		return true
	}
	return false
}

func initCommand(rootCmd *cobra.Command, config *specs.LxdComposeConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "LXD Compose configuration file")
	pflags.String("lxd-config-dir", "", "Override LXD config directory.")
	pflags.String("render-values", "", "Override render values file.")
	pflags.String("render-default", "", "Override render default file.")
	pflags.Bool("cmds-output", config.Viper.GetBool("logging.cmds_output"),
		"Show hooks commands output or not.")
	pflags.BoolP("debug", "d", config.Viper.GetBool("general.debug"),
		"Enable debug output.")

	pflags.Bool("push-progress", config.Viper.GetBool("logging.push_progressbar"),
		"Show sync files progress bar.")
	pflags.Bool("p2p-mode", config.Viper.GetBool("general.p2pmode"),
		"Enable/Disable p2p mode.")
	pflags.Bool("legacy-api", config.Viper.GetBool("general.legacyapi"),
		"Uses legacy API for Containers.")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("render_default_file", pflags.Lookup("render-default"))
	config.Viper.BindPFlag("render_values_file", pflags.Lookup("render-values"))
	config.Viper.BindPFlag("general.debug", pflags.Lookup("debug"))
	config.Viper.BindPFlag("general.p2pmode", pflags.Lookup("p2p-mode"))
	config.Viper.BindPFlag("general.legacyapi", pflags.Lookup("legacy-api"))
	config.Viper.BindPFlag("general.lxd_confdir", pflags.Lookup("lxd-config-dir"))
	config.Viper.BindPFlag("logging.cmds_output", pflags.Lookup("cmds-output"))
	config.Viper.BindPFlag("logging.push_progressbar", pflags.Lookup("push-progress"))

	rootCmd.AddCommand(
		newAclCommand(config),
		newApplyCommand(config),
		newBackupCommand(config),
		newGroupCommand(config),
		newDestroyCommand(config),
		newPrintCommand(config),
		newValidateCommand(config),
		newCompileCommand(config),
		newImagesCommand(config),
		newNodeCommand(config),
		newNetworkCommand(config),
		newStorageCommand(config),
		newPackCommand(config),
		newUnpackCommand(config),
		newProfileCommand(config),
		newDiagnoseCommand(config),
		newProjectCommand(config),
		newCommandCommand(config),
		newFetchCommand(config),
		newStopCommand(config),
	)
}

func version() string {
	ans := fmt.Sprintf("%s-g%s %s", LXD_COMPOSE_VERSION, BuildCommit, BuildTime)
	if BuildGoVersion != "" {
		ans += " " + BuildGoVersion
	}
	return ans
}

func Execute() {
	// Create Main Instance Config object
	var config *specs.LxdComposeConfig = specs.NewLxdComposeConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      version(),
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
			if err != nil && cmdNeedConfig(cmd.CalledAs()) {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
