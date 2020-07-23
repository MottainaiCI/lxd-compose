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
	LXD_COMPOSE_CONFIGNAME = ".lxd-compose"
	LXD_COMPOSE_ENV_PREFIX = "LXD_COMPOSE"
)

type LxdComposeConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General         LxdCGeneral `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	EnvironmentDirs []string    `mapstructure:"env_dirs,omitempty" json:"env_dirs,omitempty" yaml:"env_dirs,omitempty"`
}

type LxdCGeneral struct {
	Debug      bool   `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
	LxdConfDir string `mapstructure:"lxd_confdir,omitempty" yaml:"lxd_confdir,omitempty"`
}

func NewLxdComposeConfig(viper *v.Viper) *LxdComposeConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &LxdComposeConfig{Viper: viper}
}

func (c *LxdComposeConfig) GetGeneral() *LxdCGeneral {
	return &c.General
}

func (c *LxdComposeConfig) GetEnvironmentDirs() []string {
	return c.EnvironmentDirs
}

func (c *LxdComposeConfig) Unmarshal() error {
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

func (c *LxdComposeConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)
	viper.SetDefault("general.lxd_confdir", "")
	viper.SetDefault("env_dirs", []string{"./lxd-compose/envs"})
}

func (g *LxdCGeneral) HasDebug() bool {
	return g.Debug
}
