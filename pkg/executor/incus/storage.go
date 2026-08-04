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

func (e *IncusExecutor) GetStorageList() ([]string, error) {
	return e.Client.GetStoragePoolNames()
}

func (e *IncusExecutor) IsPresentStorage(name string) (bool, error) {
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

func (e *IncusExecutor) CreateStorage(sto specs.LxdCStorage) error {
	if sto.Name == "" {
		return errors.New("Invalid storage with empty name")
	}

	lxdStorage := incus_api.StoragePoolsPost{
		Name:   sto.Name,
		Driver: sto.Driver,
		StoragePoolPut: incus_api.StoragePoolPut{
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

	return e.Client.CreateStoragePool(lxdStorage)
}

func (e *IncusExecutor) UpdateStorage(sto specs.LxdCStorage) error {
	if sto.Name == "" {
		return errors.New("Invalid storage with empty name")
	}

	lxdStoragePut := incus_api.StoragePoolPut{
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

	return e.Client.UpdateStoragePool(sto.Name, lxdStoragePut, "")
}
