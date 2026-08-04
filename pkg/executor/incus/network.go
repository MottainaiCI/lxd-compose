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

func (e *IncusExecutor) GetNetworkList() ([]string, error) {
	return e.Client.GetNetworkNames()
}

func (e *IncusExecutor) IsPresentNetwork(name string) (bool, error) {
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

func (e *IncusExecutor) CreateNetwork(net specs.LxdCNetwork) error {
	if net.Name == "" {
		return errors.New("Invalid network with empty name")
	}

	lxdNetwork := incus_api.NetworksPost{
		Name: net.Name,
		Type: net.Type,
		NetworkPut: incus_api.NetworkPut{
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

	return e.Client.CreateNetwork(lxdNetwork)
}

func (e *IncusExecutor) UpdateNetwork(net specs.LxdCNetwork) error {
	if net.Name == "" {
		return errors.New("Invalid network with empty name")
	}

	lxdNetworkPut := incus_api.NetworkPut{
		Config:      net.Config,
		Description: net.Description,
	}

	if lxdNetworkPut.Config == nil {
		lxdNetworkPut.Config = make(map[string]string, 0)
	}
	if lxdNetworkPut.Description == "" {
		lxdNetworkPut.Description = fmt.Sprintf("Network %s created by lxd-compose", net.Name)
	}

	return e.Client.UpdateNetwork(net.Name, lxdNetworkPut, "")
}

func (e *IncusExecutor) SyncNetworkForwarders(net *specs.LxdCNetwork) error {
	if net.Name == "" {
		return errors.New("Invalid network with empty name")
	}

	// Retrieve the list of the NetworkForwards
	listenAddresses, err := e.Client.GetNetworkForwardAddresses(net.Name)
	if err != nil {
		return errors.New("Error on retrieve list of forwarders: " + err.Error())
	}

	if len(net.Forwards) == 0 && len(listenAddresses) == 0 {
		// POST: nothing to do
		return nil
	}

	laMap := make(map[string]bool, 0)
	// Check if there are listenAddress to remove
	for _, la := range listenAddresses {
		if !net.IsPresentForwardAddress(la) {
			err = e.Client.DeleteNetworkForward(net.Name, la)
			if err != nil {
				return errors.New(
					fmt.Sprintf(
						"Error on delete network forward for listen address %s: %s",
						la, err.Error()),
				)
			}
		} else {
			laMap[la] = true
		}
	}

	// Create or update the available listenAddresses
	for idx := range net.Forwards {

		_, toUpdate := laMap[net.Forwards[idx].ListenAddress]

		if toUpdate {
			put := e.netForward2Lxd(&net.Forwards[idx])
			err = e.Client.UpdateNetworkForward(
				net.Name,
				net.Forwards[idx].ListenAddress,
				*put, "",
			)
			if err != nil {
				return errors.New(fmt.Sprintf(
					"Error on update net forward %s: %s",
					net.Forwards[idx].ListenAddress,
					err.Error()))
			}

		} else {
			// POST: new Listen Address

			put := e.netForward2Lxd(&net.Forwards[idx])
			post := incus_api.NetworkForwardsPost{
				NetworkForwardPut: *put,
				ListenAddress:     net.Forwards[idx].ListenAddress,
			}

			err := e.Client.CreateNetworkForward(
				net.Name, post,
			)
			if err != nil {
				return errors.New(fmt.Sprintf(
					"Error on create net forward %s: %s",
					net.Forwards[idx].ListenAddress,
					err.Error()))
			}
		}
	}

	return nil
}

func (e *IncusExecutor) netForward2Lxd(f *specs.LxdCNetworkForward) *incus_api.NetworkForwardPut {
	ans := &incus_api.NetworkForwardPut{
		Description: f.Description,
		Config:      f.Config,
		Ports:       []incus_api.NetworkForwardPort{},
	}

	if ans.Config == nil {
		ans.Config = make(map[string]string, 0)
	}
	if ans.Description == "" {
		ans.Description = fmt.Sprintf(
			"Network forward for ip %s created by lxd-compose",
			f.ListenAddress,
		)
	}

	if len(f.Ports) > 0 {
		for idx := range f.Ports {
			ans.Ports = append(ans.Ports,
				incus_api.NetworkForwardPort{
					Description:   f.Ports[idx].Description,
					Protocol:      f.Ports[idx].Protocol,
					ListenPort:    f.Ports[idx].ListenPort,
					TargetPort:    f.Ports[idx].TargetPort,
					TargetAddress: f.Ports[idx].TargetAddress,
				},
			)
		}
	}

	return ans
}
