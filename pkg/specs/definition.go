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
	"gopkg.in/yaml.v2"
)

type LxdCEnvironment struct {
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	File    string `json:"-" yaml:"-"`

	TemplateEngine LxdCTemplateEngine `json:"template_engine,omitempty" yaml:"template_engine,omitempty"`

	Projects []LxdCProject `json:"projects" yaml:"projects"`
}

type LxdCTemplateEngine struct {
	Engine string `json:"engine" yaml:"engine"`
	Opts   string `json:"opts,omitempty" yaml:"opts,omitempty"`
}

type LxdCProject struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	IncludeGroupFiles []string `json:"include_groups_files,omitempty" yaml:"include_groups_files,omitempty"`
	IncludeEnvFiles   []string `json:"include_env_files,omitempty" yaml:"include_env_files,omitempty"`

	Environments []LxdCEnvVars `json:"vars,omitempty" yaml:"vars,omitempty"`

	Groups []LxdCGroup `json:"groups" yaml:"groups"`
}

type LxdCGroup struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Connection  string `json:"connection,omitempty" yaml:"connection,omitempty"`

	CommonProfiles []string `json:"common_profiles,omitempty" yaml:"common_profiles,omitempty"`
	Ephemeral      bool     `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`

	ImageFetchOptions []string `json:"image_fetch_opts,omitempty" yaml:"image_fetch_opts,omitempty"`

	Nodes []LxdCNode `json:"nodes" yaml:"nodes"`
}

type LxdCEnvVars struct {
	EnvVars map[string]interface{} `json:"envs,omitempty" yaml:"envs,omitempty"`
}

type LxdCNode struct {
	Name              string `json:"name" yaml:"name"`
	ImageSource       string `json:"image_source" yaml:"image_source"`
	ImageRemoteServer string `json:"image_remote_server,omitempty" yaml:"image_remote_server,omitempty"`

	Labels []string `json:"labels,omitempty" yaml:"labels,omitempty"`

	SourceDir string `json:"source_dir,omitempty" yaml:"source_dir,omitempty"`

	Entrypoint       string   `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	BootstrapCommand []string `json:"bootstrap_commands,omitempty" yaml:"bootstrap_commands,omitempty"`
	SyncPostCommands []string `json:"sync_post_commands,omitempty" yaml:"sync_post_commands,omitempty"`

	ConfigTemplates []LxdCConfigTemplate `json:"config_templates,omitempty" yaml:"config_templates,omitempty"`
	SyncResources   []LxdCSyncResource   `json:"sync_resources,omitempty" yaml:"sync_resources,omitempty"`
}

type LxdCConfigTemplate struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"dst" yaml:"dst"`
}

type LxdCSyncResource struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"dst" yaml:"dst"`
	Recursive   bool   `json:"recursive,omitempty" yaml:"recursive,omitempty"`
}

func GroupFromYaml(data []byte) (*LxdCGroup, error) {
	ans := &LxdCGroup{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	return ans, nil
}

func EnvVarsFromYaml(data []byte) (*LxdCEnvVars, error) {
	ans := &LxdCEnvVars{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	return ans, nil
}

func EnvironmentFromYaml(data []byte, file string) (*LxdCEnvironment, error) {
	ans := &LxdCEnvironment{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file
	return ans, nil
}

func (p *LxdCProject) AddGroup(grp *LxdCGroup) {
	p.Groups = append(p.Groups, *grp)
}

func (p *LxdCProject) AddEnvironment(e *LxdCEnvVars) {
	p.Environments = append(p.Environments, *e)
}
