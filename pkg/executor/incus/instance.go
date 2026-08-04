/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	base "github.com/MottainaiCI/lxd-compose/pkg/executor/base"

	incus_api "github.com/lxc/incus/v7/shared/api"
)

// Get instance data and the ETag
func (e *IncusExecutor) GetInstance(name string) (*incus_api.Instance, string, error) {
	return e.Client.GetInstance(name)
}

func (e *IncusExecutor) UpdateInstance(
	name string, idata *incus_api.InstancePut,
	etag string) error {

	oper, err := e.Client.UpdateInstance(name, *idata, etag)
	if err != nil {
		return err
	}

	err = e.WaitOperation(oper, nil)
	if err != nil {
		return err
	}

	e.Emitter.Emits(base.LxdContainerUpdated, map[string]interface{}{
		"name":      name,
		"profiles":  idata.Profiles,
		"ephemeral": idata.Ephemeral,
		"config":    idata.Config,
		"devices":   idata.Devices,
	})

	return nil
}

func (e *IncusExecutor) RemoveProfilesFromInstance(name string, profiles []string) error {
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
	for _, p := range idata.Profiles {
		if _, present := mprofiles[p]; !present {
			newProfilesList = append(newProfilesList, p)
		}
	}

	iput := &incus_api.InstancePut{
		Profiles: newProfilesList,
	}

	err = e.UpdateInstance(name, iput, etag)
	if err != nil {
		return err
	}

	return nil
}

func (e *IncusExecutor) AddProfiles2Instance(name string, profiles []string) error {
	// Retrieve the current status of the instance
	idata, etag, err := e.GetInstance(name)
	if err != nil {
		return err
	}

	// Convert profiles to add in map
	mprofiles := make(map[string]bool, 0)
	for _, p := range idata.Profiles {
		mprofiles[p] = true
	}

	// Check if the profiles to add are present
	update2do := false
	newProfilesList := idata.Profiles
	for _, p := range profiles {
		if _, present := mprofiles[p]; !present {
			update2do = true
			newProfilesList = append(newProfilesList, p)
		}
	}

	if update2do {
		iput := &incus_api.InstancePut{
			Profiles: newProfilesList,
		}

		err := e.UpdateInstance(name, iput, etag)
		if err != nil {
			return err
		}
	}

	return nil
}
