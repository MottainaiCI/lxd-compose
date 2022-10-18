/*
Copyright (C) 2020-2022  Daniele Rondina <geaaru@funtoo.org>
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

func (e *LxdCExecutor) GetNetworkList() ([]string, error) {
	return e.LxdClient.GetNetworkNames()
}

func (e *LxdCExecutor) IsPresentNetwork(name string) (bool, error) {
	ans := false
	list, err := e.GetNetworkList()

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

func (e *LxdCExecutor) CreateNetwork(net specs.LxdCNetwork) error {
	if net.Name == "" {
		return errors.New("Invalid network with empty name")
	}

	lxdNetwork := lxd_api.NetworksPost{
		Name: net.Name,
		Type: net.Type,
		NetworkPut: lxd_api.NetworkPut{
			Config:      net.Config,
			Description: net.Description,
		},
	}

	if lxdNetwork.NetworkPut.Config == nil {
		lxdNetwork.NetworkPut.Config = make(map[string]string, 0)
	}
	if lxdNetwork.NetworkPut.Description == "" {
		lxdNetwork.NetworkPut.Description = fmt.Sprintf("Network %s created by lxd-compose", net.Name)
	}

	return e.LxdClient.CreateNetwork(lxdNetwork)
}

func (e *LxdCExecutor) UpdateNetwork(net specs.LxdCNetwork) error {
	if net.Name == "" {
		return errors.New("Invalid network with empty name")
	}

	lxdNetworkPut := lxd_api.NetworkPut{
		Config:      net.Config,
		Description: net.Description,
	}

	if lxdNetworkPut.Config == nil {
		lxdNetworkPut.Config = make(map[string]string, 0)
	}
	if lxdNetworkPut.Description == "" {
		lxdNetworkPut.Description = fmt.Sprintf("Network %s created by lxd-compose", net.Name)
	}

	return e.LxdClient.UpdateNetwork(net.Name, lxdNetworkPut, "")
}
