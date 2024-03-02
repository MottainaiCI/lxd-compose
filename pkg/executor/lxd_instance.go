/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	lxd_api "github.com/canonical/lxd/shared/api"
)

// Get instance data and the ETag
func (e *LxdCExecutor) GetInstance(name string) (*lxd_api.Instance, string, error) {
	return e.LxdClient.GetInstance(name)
}

func (e *LxdCExecutor) UpdateInstance(
	name string, idata *lxd_api.InstancePut,
	etag string) error {

	oper, err := e.LxdClient.UpdateInstance(name, *idata, etag)
	if err != nil {
		return err
	}

	err = e.WaitOperation(oper, nil)
	if err != nil {
		return err
	}

	e.Emitter.Emits(LxdContainerUpdated, map[string]interface{}{
		"name":      name,
		"profiles":  idata.Profiles,
		"ephemeral": idata.Ephemeral,
		"config":    idata.Config,
		"devices":   idata.Devices,
	})

	return nil
}

func (e *LxdCExecutor) RemoveProfilesFromInstance(name string, profiles []string) error {
	// Retrieve the current status of the instance
	idata, etag, err := e.GetInstance(name)
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
	for _, p := range idata.InstancePut.Profiles {
		if _, present := mprofiles[p]; !present {
			newProfilesList = append(newProfilesList, p)
		}
	}

	idata.InstancePut.Profiles = newProfilesList

	err = e.UpdateInstance(name, &idata.InstancePut, etag)
	if err != nil {
		return err
	}

	return nil
}

func (e *LxdCExecutor) AddProfiles2Instance(name string, profiles []string) error {
	// Retrieve the current status of the instance
	idata, etag, err := e.GetInstance(name)
	if err != nil {
		return err
	}

	// Convert profiles to add in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range idata.InstancePut.Profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to add are present
	update2do := false
	newProfilesList := idata.InstancePut.Profiles
	for _, p := range profiles {
		if _, present := mprofiles[p]; !present {
			update2do = true
			newProfilesList = append(newProfilesList, p)
		}
	}

	if update2do {
		idata.InstancePut.Profiles = newProfilesList

		err := e.UpdateInstance(name, &idata.InstancePut, etag)
		if err != nil {
			return err
		}
	}

	return nil
}
