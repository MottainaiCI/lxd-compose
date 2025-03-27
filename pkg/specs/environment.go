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
package specs

import (
	"errors"
	"path"

	"gopkg.in/yaml.v3"
)

func EnvironmentFromYaml(data []byte, file string) (*LxdCEnvironment, error) {
	ans := &LxdCEnvironment{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	if ans.Commands == nil {
		ans.Commands = []LxdCCommand{}
	}
	if ans.IncludeCommandsFiles == nil {
		ans.IncludeCommandsFiles = []string{}
	}

	for idx := range ans.Projects {
		ans.Projects[idx].Init()
	}

	return ans, nil
}

func (e *LxdCEnvironment) GetProjectByName(pName string) *LxdCProject {
	for idx, p := range e.Projects {
		if p.Name == pName {
			return &e.Projects[idx]
		}
	}

	return nil
}

func (e *LxdCEnvironment) GetProjects() *[]LxdCProject {
	return &e.Projects
}

func (e *LxdCEnvironment) GetProfiles() *[]LxdCProfile {
	return &e.Profiles
}

func (e *LxdCEnvironment) GetCommands() *[]LxdCCommand {
	return &e.Commands
}

func (e *LxdCEnvironment) GetCommand(name string) (*LxdCCommand, error) {
	for idx, cmd := range e.Commands {
		if cmd.Name == name {
			return &e.Commands[idx], nil
		}
	}

	return nil, errors.New("Command + " + name + " not available.")
}

func (e *LxdCEnvironment) AddCommand(cmd *LxdCCommand) {
	e.Commands = append(e.Commands, *cmd)
}

func (e *LxdCEnvironment) GetProfile(name string) (LxdCProfile, error) {
	ans := LxdCProfile{}

	for _, prof := range e.Profiles {
		if prof.Name == name {
			return prof, nil
		}
	}

	return ans, errors.New("Profile " + name + " not available.")
}

func (e *LxdCEnvironment) GetNetworks() *[]LxdCNetwork {
	return &e.Networks
}

func (e *LxdCEnvironment) GetStorages() *[]LxdCStorage {
	return &e.Storages
}

func (e *LxdCEnvironment) GetACLs() *[]LxdCAcl {
	return &e.Acls
}

func (e *LxdCEnvironment) GetACL(name string) (LxdCAcl, error) {
	ans := LxdCAcl{}

	for _, acl := range e.Acls {
		if acl.Name == name {
			return acl, nil
		}
	}

	return ans, errors.New("ACL " + name + " not available.")
}

func (e *LxdCEnvironment) GetNetwork(name string) (LxdCNetwork, error) {
	ans := LxdCNetwork{}

	for _, net := range e.Networks {
		if net.Name == name {
			return net, nil
		}
	}

	return ans, errors.New("Network " + name + " not available.")
}

func (e *LxdCEnvironment) GetStorage(name string) (LxdCStorage, error) {
	ans := LxdCStorage{}

	for _, st := range e.Storages {
		if st.Name == name {
			return st, nil
		}
	}

	return ans, errors.New("Storage " + name + " not available.")
}

func (e *LxdCEnvironment) AddNetwork(network *LxdCNetwork) {
	e.Networks = append(e.Networks, *network)
}

func (e *LxdCEnvironment) AddStorage(storage *LxdCStorage) {
	e.Storages = append(e.Storages, *storage)
}

func (e *LxdCEnvironment) AddProfile(profile *LxdCProfile) {
	e.Profiles = append(e.Profiles, *profile)
}

func (e *LxdCEnvironment) AddACL(acl *LxdCAcl) {
	e.Acls = append(e.Acls, *acl)
}

func (e *LxdCEnvironment) GetBaseFile() string {
	ans := ""
	if e.File != "" {
		ans = path.Base(e.File)
	}

	return ans
}
