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

func (e *LxdCExecutor) GetStorageList() ([]string, error) {
	return e.LxdClient.GetStoragePoolNames()
}

func (e *LxdCExecutor) IsPresentStorage(name string) (bool, error) {
	ans := false
	list, err := e.GetStorageList()

	if err != nil {
		return false, err
	}

	for _, n := range list {
		if n == name {
			ans = true
			break
		}
	}

	return ans, nil
}

func (e *LxdCExecutor) CreateStorage(sto specs.LxdCStorage) error {
	if sto.Name == "" {
		return errors.New("Invalid storage with empty name")
	}

	lxdStorage := lxd_api.StoragePoolsPost{
		Name:   sto.Name,
		Driver: sto.Driver,
		StoragePoolPut: lxd_api.StoragePoolPut{
			Config:      sto.Config,
			Description: sto.Description,
		},
	}

	if lxdStorage.StoragePoolPut.Config == nil {
		lxdStorage.StoragePoolPut.Config = make(map[string]string, 0)
	}
	if lxdStorage.StoragePoolPut.Description == "" {
		lxdStorage.StoragePoolPut.Description = fmt.Sprintf(
			"Storage %s created by lxd-compose",
			sto.Name,
		)
	}

	return e.LxdClient.CreateStoragePool(lxdStorage)
}

func (e *LxdCExecutor) UpdateStorage(sto specs.LxdCStorage) error {
	if sto.Name == "" {
		return errors.New("Invalid storage with empty name")
	}

	lxdStoragePut := lxd_api.StoragePoolPut{
		Config:      sto.Config,
		Description: sto.Description,
	}

	if lxdStoragePut.Config == nil {
		lxdStoragePut.Config = make(map[string]string, 0)
	}
	if lxdStoragePut.Description == "" {
		lxdStoragePut.Description = fmt.Sprintf(
			"Storage %s created by lxd-compose",
			sto.Name,
		)
	}

	return e.LxdClient.UpdateStoragePool(sto.Name, lxdStoragePut, "")
}
