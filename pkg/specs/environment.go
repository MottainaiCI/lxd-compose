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
	"errors"

	"gopkg.in/yaml.v2"
)

func EnvironmentFromYaml(data []byte, file string) (*LxdCEnvironment, error) {
	ans := &LxdCEnvironment{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}
	ans.File = file

	for idx, _ := range ans.Projects {
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

func (e *LxdCEnvironment) GetProfile(name string) (LxdCProfile, error) {
	ans := LxdCProfile{}

	for _, prof := range e.Profiles {
		if prof.Name == name {
			return prof, nil
		}
	}

	return ans, errors.New("Profile " + name + " not available.")
}
