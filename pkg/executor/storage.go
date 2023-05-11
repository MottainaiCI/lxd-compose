/*
Copyright (C) 2020-2021  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"errors"
	"fmt"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd_api "github.com/lxc/lxd/shared/api"
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
