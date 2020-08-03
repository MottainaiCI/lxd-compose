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
package specs

import (
	v "github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

const (
	TM_ENV_PREFIX = "TM"
	TM_CONFIGNAME = ".time-master"
)

type TimeMasterConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General TimeMasterConfigGeneral `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging TimeMasterConfigLogging `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`

	Work TimeMasterConfigWork `mapstructure:"work,omitempty" json:"work,omitempty" yaml:"work,omitempty"`

	ClientsDirs []string `mapstructure:"clients_dirs,omitempty" json:"clients_dirs,omitempty" yaml:"clients_dirs,omitempty"`

	ResourcesDirs []string `mapstructure:"resources_dirs,omitempty" json:"resources_dirs,omitempty" yaml:"resources_dirs,omitempty"`

	ScenariosDirs []string `mapstructure:"scenarios_dirs,omitempty" json:"scenarios_dirs,omitempty" yaml:"scenarios_dirs,omitempty"`
}

type TimeMasterConfigGeneral struct {
	Debug bool `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
}

type TimeMasterConfigLogging struct {
	// Path of the logfile
	Path string `mapstructure:"path,omitempty" json:"path,omitempty" yaml:"path,omitempty"`
	// Enable/Disable logging to file
	EnableLogFile bool `mapstructure:"enable_logfile,omitempty" json:"enable_logfile,omitempty" yaml:"enable_logfile,omitempty"`
	// Enable JSON format logging in file
	JsonFormat bool `mapstructure:"json_format,omitempty" json:"json_format,omitempty" yaml:"json_format,omitempty"`

	// Log level
	Level string `mapstructure:"level,omitempty" json:"level,omitempty" yaml:"level,omitempty"`

	// Enable emoji
	EnableEmoji bool `mapstructure:"enable_emoji,omitempty" json:"enable_emoji,omitempty" yaml:"enable_emoji,omitempty"`
	// Enable/Disable color in logging
	Color bool `mapstructure:"color,omitempty" json:"color,omitempty" yaml:"color,omitempty"`
}

type TimeMasterConfigWork struct {
	// Default number of hours for day
	WorkHours int `mapstructure:"work_hours,omitempty" json:"work_hours,omitempty" yaml:"work_hours,omitempty"`
}

func NewTimeMasterConfig(viper *v.Viper) *TimeMasterConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &TimeMasterConfig{Viper: viper}
}

func (c *TimeMasterConfig) GetWork() *TimeMasterConfigWork {
	return &c.Work
}

func (c *TimeMasterConfig) GetGeneral() *TimeMasterConfigGeneral {
	return &c.General
}

func (c *TimeMasterConfig) GetLogging() *TimeMasterConfigLogging {
	return &c.Logging
}

func (c *TimeMasterConfig) GetClientsDirs() []string {
	return c.ClientsDirs
}

func (c *TimeMasterConfig) GetResourcesDirs() []string {
	return c.ResourcesDirs
}

func (c *TimeMasterConfig) GetScenariosDirs() []string {
	return c.ScenariosDirs
}

func (c *TimeMasterConfig) Unmarshal() error {
	var err error

	if c.Viper.InConfig("etcd-config") &&
		c.Viper.GetBool("etcd-config") {
		err = c.Viper.ReadRemoteConfig()
	} else {
		err = c.Viper.ReadInConfig()
	}

	err = c.Viper.Unmarshal(&c)

	return err
}

func (c *TimeMasterConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)

	viper.SetDefault("work.work_hours", 8)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "/var/log/luet.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)

	viper.SetDefault("clients_dirs", []string{"./clients"})
	viper.SetDefault("resources_dirs", []string{"./resources"})
	viper.SetDefault("scenarios_dirs", []string{"./scenarios"})
}

func (g *TimeMasterConfigGeneral) HasDebug() bool {
	return g.Debug
}
