/*
Copyright (C) 2020-2024  Daniele Rondina <geaaru@gmail.com>
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
package executor

import (
	lxd_api "github.com/canonical/lxd/shared/api"
)

// Get the container data and the ETag
func (e *LxdCExecutor) GetContainer(name string) (*lxd_api.Container, string, error) {
	return e.LxdClient.GetContainer(name)
}

func (e *LxdCExecutor) UpdateContainer(name string, cdata *lxd_api.ContainerPut, etag string) error {
	oper, err := e.LxdClient.UpdateContainer(name, *cdata, etag)
	if err != nil {
		return err
	}

	err = e.WaitOperation(oper, nil)
	if err != nil {
		return err
	}

	e.Emitter.Emits(LxdContainerUpdated, map[string]interface{}{
		"name":      name,
		"profiles":  cdata.Profiles,
		"ephemeral": cdata.Ephemeral,
		"config":    cdata.Config,
		"devices":   cdata.Devices,
	})

	return nil
}

func (e *LxdCExecutor) RemoveProfilesFromContainer(name string, profiles []string) error {
	// Retrieve the current status of the container
	cdata, etag, err := e.GetContainer(name)
	if err != nil {
		return err
	}

	// Convert profiles to remove in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to remove are present
	newProfilesList := []string{}
	for _, p := range cdata.ContainerPut.Profiles {
		if _, present := mprofiles[p]; !present {
			newProfilesList = append(newProfilesList, p)
		}
	}

	cdata.ContainerPut.Profiles = newProfilesList

	err = e.UpdateContainer(name, &cdata.ContainerPut, etag)
	if err != nil {
		return err
	}

	return nil
}

func (e LxdCExecutor) AddProfiles2Container(name string, profiles []string) error {
	// Retrieve the current status of the container
	cdata, etag, err := e.GetContainer(name)
	if err != nil {
		return err
	}

	// Convert profiles to add in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range cdata.ContainerPut.Profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to add are present
	update2do := false
	newProfilesList := cdata.ContainerPut.Profiles
	for _, p := range profiles {
		if _, present := mprofiles[p]; !present {
			update2do = true
			newProfilesList = append(newProfilesList, p)
		}
	}

	if update2do {
		cdata.ContainerPut.Profiles = newProfilesList

		err := e.UpdateContainer(name, &cdata.ContainerPut, etag)
		if err != nil {
			return err
		}
	}

	return nil
}
