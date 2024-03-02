/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd_api "github.com/canonical/lxd/shared/api"
)

func (e *LxdCExecutor) GetProfilesList() ([]string, error) {
	return e.LxdClient.GetProfileNames()
}

func (e *LxdCExecutor) IsPresentProfile(profileName string) (bool, error) {
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

func (e *LxdCExecutor) CreateProfile(profile specs.LxdCProfile) error {
	if profile.Name == "" {
		return errors.New("Invalid profile with empty name")
	}

	lxdProfile := lxd_api.ProfilesPost{
		ProfilePut: lxd_api.ProfilePut{
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

	return e.LxdClient.CreateProfile(lxdProfile)
}

func (e *LxdCExecutor) UpdateProfile(profile specs.LxdCProfile) error {
	if profile.Name == "" {
		return errors.New("Invalid profile with empty name")
	}

	lxdProfilePut := lxd_api.ProfilePut{
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

	return e.LxdClient.UpdateProfile(profile.Name, lxdProfilePut, "")
}
