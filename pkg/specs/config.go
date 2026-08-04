/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	helpers_sec "github.com/MottainaiCI/lxd-compose/pkg/helpers/security"

	v "github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	LXD_COMPOSE_CONFIGNAME = ".lxd-compose"
	LXD_COMPOSE_ENV_PREFIX = "LXD_COMPOSE"
)

type LxdComposeConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General         LxdCGeneral  `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging         LxdCLogging  `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`
	Security        LxdCSecurity `mapstructure:"security" json:"security,omitempty" yaml:"security,omitempty"`
	EnvironmentDirs []string     `mapstructure:"env_dirs,omitempty" json:"env_dirs,omitempty" yaml:"env_dirs,omitempty"`

	RenderDefaultFile   string                 `mapstructure:"render_default_file,omitempty" json:"render_default_file,omitempty" yaml:"render_default_file,omitempty"`
	RenderValuesFile    string                 `mapstructure:"render_values_file,omitempty" json:"render_values_file,omitempty" yaml:"render_values_file,omitempty"`
	RenderSecretFile    string                 `mapstructure:"render_secrets_file,omitempty" json:"render_secrets_file,omitempty" yaml:"render_secrets_file,omitempty"`
	RenderEnvsVars      map[string]interface{} `mapstructure:"-" json:"-" yaml:"-"`
	RenderTemplatesDirs []string               `mapstructure:"render_templates_dirs,omitempty" json:"render_templates_dirs,omitempty" yaml:"render_templates_dirs,omitempty"`
}

type LxdCGeneral struct {
	Debug           bool   `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
	LxdConfDir      string `mapstructure:"lxd_confdir,omitempty" json:"lxd_confdir,omitempty" yaml:"lxd_confdir,omitempty"`
	LxdLocalDisable bool   `mapstructure:"lxd_local_disable,omitempty" json:"lxd_local_disable,omitempty" yaml:"lxd_local_disable,omitempty"`
	P2PMode         bool   `mapstructure:"p2pmode,omitempty" json:"p2pmode,omitempty" yaml:"p2pmode,omitempty"`
	LegacyApi       bool   `mapstructure:"legacyapi,omitempty" json:"legacyapi,omitempty" yaml:"legacyapi,omitempty"`

	Concurrency int `mapstructure:"concurrency,omitempty" json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
}

type LxdCSecurity struct {
	Keyfile        string `mapstructure:"keyfile" json:"keyfile,omitempty" yaml:"keyfile,omitempty"`
	Key            string `mapstructure:"key" json:"key,omitempty" yaml:"key,omitempty"`
	EncryptSecrets *bool  `mapstructure:"encrypted_secrets" json:"encrypted_secrets,omitempty" yaml:"encrypted_secrets,omitempty"`

	DKAOpts *LxdCDKAOpts `mapstructure:"dka_opts" json:"dka_opts,omitempty" yaml:"dka_opts,omitempty"`
}

type LxdCDKAOpts struct {
	TimeIterations *uint32 `mapstructure:"time_iterations" json:"time_iterations,omitempty" yaml:"time_iterations,omitempty"`
	MemoryUsage    *uint32 `mapstructure:"memory_usage" json:"memory_usage,omitempty" yaml:"memory_usage,omitempty"`
	KeyLength      *uint32 `mapstructure:"key_length" json:"key_length,omitempty" yaml:"key_length,omitempty"`
	Parallelism    *uint8  `mapstructure:"parallelism" json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
}

type LxdCLogging struct {
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

	// Enable/Disable commands output logging
	RuntimeCmdsOutput bool `mapstructure:"runtime_cmds_output,omitempty" json:"runtime_cmds_output,omitempty" yaml:"runtime_cmds_output,omitempty"`
	CmdsOutput        bool `mapstructure:"cmds_output,omitempty" json:"cmds_output,omitempty" yaml:"cmds_output,omitempty"`
	PushProgressBar   bool `mapstructure:"push_progressbar,omitempty" json:"push_progressbar,omitempty" yaml:"push_progressbar,omitempty"`
}

func NewLxdComposeConfig(viper *v.Viper) *LxdComposeConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &LxdComposeConfig{Viper: viper}
}

func (c *LxdComposeConfig) Clone() *LxdComposeConfig {
	ans := NewLxdComposeConfig(nil)

	ans.EnvironmentDirs = c.EnvironmentDirs
	ans.RenderDefaultFile = c.RenderDefaultFile
	ans.RenderValuesFile = c.RenderValuesFile
	ans.RenderSecretFile = c.RenderSecretFile
	ans.RenderTemplatesDirs = c.RenderTemplatesDirs

	ans.General.Debug = c.General.Debug
	ans.General.LegacyApi = c.General.LegacyApi
	ans.General.LxdConfDir = c.General.LxdConfDir
	ans.General.LxdLocalDisable = c.General.LxdLocalDisable
	ans.General.P2PMode = c.General.P2PMode
	ans.General.Concurrency = c.General.Concurrency

	ans.Logging.Path = c.Logging.Path
	ans.Logging.EnableLogFile = c.Logging.EnableLogFile
	ans.Logging.JsonFormat = c.Logging.JsonFormat
	ans.Logging.Level = c.Logging.Level
	ans.Logging.EnableEmoji = c.Logging.EnableEmoji
	ans.Logging.Color = c.Logging.Color
	ans.Logging.RuntimeCmdsOutput = c.Logging.RuntimeCmdsOutput
	ans.Logging.CmdsOutput = c.Logging.CmdsOutput
	ans.Logging.PushProgressBar = c.Logging.PushProgressBar

	ans.Security.Keyfile = c.Security.Keyfile
	ans.Security.Key = c.Security.Key

	if c.Security.EncryptSecrets != nil {
		es := *c.Security.EncryptSecrets
		ans.Security.EncryptSecrets = &es
	}

	if ans.Security.DKAOpts != nil {
		ans.Security.DKAOpts = &LxdCDKAOpts{}

		if c.Security.DKAOpts.TimeIterations != nil {
			ti := *c.Security.DKAOpts.TimeIterations
			ans.Security.DKAOpts.TimeIterations = &ti
		}

		if c.Security.DKAOpts.MemoryUsage != nil {
			mu := *c.Security.DKAOpts.MemoryUsage
			ans.Security.DKAOpts.MemoryUsage = &mu
		}

		if c.Security.DKAOpts.KeyLength != nil {
			kl := *c.Security.DKAOpts.KeyLength
			ans.Security.DKAOpts.KeyLength = &kl
		}

		if c.Security.DKAOpts.Parallelism != nil {
			par := *c.Security.DKAOpts.Parallelism
			ans.Security.DKAOpts.Parallelism = &par
		}

	}

	return ans
}

func (c *LxdComposeConfig) GetGeneral() *LxdCGeneral {
	return &c.General
}

func (c *LxdComposeConfig) GetEnvironmentDirs() []string {
	return c.EnvironmentDirs
}

func (c *LxdComposeConfig) GetLogging() *LxdCLogging {
	return &c.Logging
}

func (c *LxdComposeConfig) GetSecurity() *LxdCSecurity {
	return &c.Security
}

func (c *LxdComposeConfig) IsEnableRenderEngine() bool {
	if c.RenderValuesFile != "" || c.RenderDefaultFile != "" {
		return true
	}
	return false
}

func (c *LxdComposeConfig) GetSecrets() (*map[string]interface{}, error) {
	ans := make(map[string]interface{}, 0)

	if c.RenderSecretFile != "" {
		data, err := os.ReadFile(c.RenderSecretFile)
		if err != nil {
			return nil, fmt.Errorf("error on read file %s: %s", c.RenderSecretFile, err.Error())
		}

		if c.Security.EncryptSecrets != nil && *c.Security.EncryptSecrets {

			keyBytes := []byte{}

			if c.GetSecurity().Key != "" {
				keyBytes, err = base64.StdEncoding.DecodeString(c.GetSecurity().Key)
				if err != nil {
					return nil, fmt.Errorf("error on decode base64 key: %s", err.Error())
				}
			}

			dkaOpts := helpers_sec.NewDKAOptsDefault()
			if c.GetSecurity().DKAOpts != nil {
				if c.GetSecurity().DKAOpts.TimeIterations != nil {
					dkaOpts.TimeIterations = *c.GetSecurity().DKAOpts.TimeIterations
				}
				if c.GetSecurity().DKAOpts.MemoryUsage != nil {
					dkaOpts.MemoryUsage = *c.GetSecurity().DKAOpts.MemoryUsage
				}
				if c.GetSecurity().DKAOpts.KeyLength != nil {
					dkaOpts.KeyLength = *c.GetSecurity().DKAOpts.KeyLength
				}
				if c.GetSecurity().DKAOpts.Parallelism != nil {
					dkaOpts.Parallelism = *c.GetSecurity().DKAOpts.Parallelism
				}
			}

			data, err = base64.StdEncoding.DecodeString(string(data))
			if err != nil {
				return nil, fmt.Errorf("error on decode base64 data: %s", err.Error())
			}

			decodedBytes, err := helpers_sec.Decrypt(data, keyBytes, dkaOpts)
			if err != nil {
				return nil, fmt.Errorf("error on decrypt secrets: %s", err.Error())
			}

			data = decodedBytes
		}

		if err = yaml.Unmarshal(data, &ans); err != nil {
			return nil, fmt.Errorf("error on unmarshal secrets: %s", err.Error())
		}
	}

	return &ans, nil
}

func (c *LxdComposeConfig) Unmarshal() error {
	var err error

	if c.Viper.InConfig("etcd-config") &&
		c.Viper.GetBool("etcd-config") {
		err = c.Viper.ReadRemoteConfig()
	} else {
		err = c.Viper.ReadInConfig()
	}

	if err != nil {
		return err
	}

	err = c.Viper.Unmarshal(&c)

	return err
}

func (c *LxdComposeConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *LxdComposeConfig) SetRenderEnvs(envs []string) error {
	e := NewEnvVars()

	for _, env := range envs {
		err := e.AddKVAggregated(env)
		if err != nil {
			return err
		}
	}

	if len(e.EnvVars) > 0 {
		c.RenderEnvsVars = e.EnvVars
	}

	return nil
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)
	viper.SetDefault("general.p2pmode", false)
	viper.SetDefault("general.legacyapi", false)
	viper.SetDefault("general.lxd_local_disable", false)
	viper.SetDefault("general.lxd_confdir", "")
	viper.SetDefault("general.concurrency", runtime.NumCPU())
	viper.SetDefault("render_default_file", "")
	viper.SetDefault("render_values_file", "")
	viper.SetDefault("render_secret_file", "")
	viper.SetDefault("render_templates_dirs", []string{})

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "./logs/lxd-compose.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)
	viper.SetDefault("logging.cmds_output", true)
	viper.SetDefault("logging.runtime_cmds_output", true)
	viper.SetDefault("logging.push_progressbar", false)

	viper.SetDefault("env_dirs", []string{"./lxd-compose/envs"})
}

func (g *LxdCGeneral) HasDebug() bool {
	return g.Debug
}
