/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	incus_api "github.com/lxc/incus/v7/shared/api"
)

func (e *IncusExecutor) GetProfilesList() ([]string, error) {
	return e.Client.GetProfileNames()
}

func (e *IncusExecutor) IsPresentProfile(profileName string) (bool, error) {
	ans := false
	list, err := e.GetProfilesList()

	if err != nil {
		return false, err
	}

	for _, p := range list {
		if p == profileName {
			ans = true
			break
		}
	}

	return ans, nil
}

func (e *IncusExecutor) CreateProfile(profile specs.LxdCProfile) error {
	if profile.Name == "" {
		return errors.New("Invalid profile with empty name")
	}

	lxdProfile := incus_api.ProfilesPost{
		ProfilePut: incus_api.ProfilePut{
			Config:      profile.Config,
			Devices:     profile.Devices,
			Description: profile.Description,
		},
		Name: profile.Name,
	}

	if lxdProfile.ProfilePut.Config == nil {
		lxdProfile.ProfilePut.Config = make(map[string]string, 0)
	}
	if lxdProfile.ProfilePut.Devices == nil {
		lxdProfile.ProfilePut.Devices = make(map[string]map[string]string, 0)
	}

	if lxdProfile.ProfilePut.Description == "" {
		lxdProfile.ProfilePut.Description =
			fmt.Sprintf("Profile %s created by lxd-compose", profile.Name)
	}

	return e.Client.CreateProfile(lxdProfile)
}

func (e *IncusExecutor) UpdateProfile(profile specs.LxdCProfile) error {
	if profile.Name == "" {
		return errors.New("Invalid profile with empty name")
	}

	lxdProfilePut := incus_api.ProfilePut{
		Config:  profile.Config,
		Devices: profile.Devices,
	}

	if profile.Description != "" {
		lxdProfilePut.Description = profile.Description
	}

	if lxdProfilePut.Config == nil {
		lxdProfilePut.Config = make(map[string]string, 0)
	}
	if lxdProfilePut.Devices == nil {
		lxdProfilePut.Devices = make(map[string]map[string]string, 0)
	}

	return e.Client.UpdateProfile(profile.Name, lxdProfilePut, "")
}
