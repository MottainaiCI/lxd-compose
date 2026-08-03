/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package lxd

import (
	"errors"
	"fmt"
	"time"

	base "github.com/MottainaiCI/lxd-compose/pkg/executor/base"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	"github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd "github.com/canonical/lxd/client"
	lxd_config "github.com/canonical/lxd/lxc/config"
	lxd_api "github.com/canonical/lxd/shared/api"
	lxd_cli "github.com/canonical/lxd/shared/cmd"
)

type LxdExecutor struct {
	*base.BaseExecutor

	LxdClient lxd.InstanceServer
	LxdConfig *lxd_config.Config
}

func NewLxdExecutorWithEmitter(endpoint, configdir string,
	entrypoint []string, ephemeral, showCmdsOutput,
	runtimeCmdsOutput bool, emitter base.LxdCExecutorEmitter) *LxdExecutor {
	return &LxdExecutor{
		BaseExecutor: base.NewBaseExecutorWithEmitter(
			endpoint, configdir, entrypoint,
			ephemeral, showCmdsOutput,
			runtimeCmdsOutput, emitter),
	}
}

func (e *LxdExecutor) GetType() string { return specs.ConnectionLxd6 }

func (e *LxdExecutor) CreateContainer(name, fingerprint, imageServer string, profiles []string) error {
	return e.CreateContainerWithConfig(name, fingerprint, imageServer, profiles, map[string]string{})
}

func (e *LxdExecutor) CreateContainerWithConfig(name, fingerprint, imageServer string, profiles []string, configMap map[string]string) error {
	if name == "" {
		return errors.New("Invalid container name")
	}

	// Check if container is already present.
	isPresent, err := e.IsPresentContainer(name)
	if err != nil {
		return err
	}

	logger := log.GetDefaultLogger()

	if isPresent {
		e.Emitter.InfoLog(false, logger.Aurora.Bold(logger.Aurora.BrightCyan(
			">>> Container "+name+" already present. Nothing to do. - :check_mark:")))
		return nil
	}

	// Pull image
	imageFingerprint, err := e.PullImage(fingerprint, imageServer)
	if err != nil {
		logger.Error("Error on pull image " + fingerprint + " from remote " + imageServer)
		return err
	}

	e.Emitter.InfoLog(true, logger.Aurora.Bold(logger.Aurora.BrightCyan(
		">>> Creating container "+name+"... - :factory:")))
	err = e.LaunchContainerWithConfig(name, imageFingerprint, profiles, configMap)
	if err != nil {
		logger.Error("Creating container error: " + err.Error())
		return err
	}

	return nil
}

func (e *LxdExecutor) StopContainer(name string) error {
	return e.DoAction2Container(name, "stop")
}

func (e *LxdExecutor) StartContainer(name string) error {
	return e.DoAction2Container(name, "start")
}

func (e *LxdExecutor) GetContainerList() ([]string, error) {
	return e.LxdClient.GetInstanceNames(lxd_api.InstanceTypeContainer)
}

func (e *LxdExecutor) IsRunningContainer(name string) (bool, error) {
	ans := false
	var status string

	iInfo, _, err := e.LxdClient.GetInstance(name)
	if err != nil {
		return ans, err
	}
	status = iInfo.Status

	if status == "Running" {
		ans = true
	}

	return ans, nil
}

func (e *LxdExecutor) IsEphemeralContainer(containerName string) (bool, error) {
	iInfo, _, err := e.LxdClient.GetInstance(containerName)
	if err != nil {
		return false, err
	}
	return iInfo.Ephemeral, nil
}

func (e *LxdExecutor) IsPresentContainer(containerName string) (bool, error) {
	ans := false
	list, err := e.GetContainerList()

	if err != nil {
		return false, err
	}

	for _, c := range list {
		if c == containerName {
			ans = true
			break
		}
	}

	return ans, nil
}

func (e *LxdExecutor) CopyContainerOnInstance(
	containerName, newContainerName string) error {

	args := lxd.InstanceCopyArgs{
		Name: newContainerName,
		// Always follow stateless copy.
		Live: false,
		// Ignore containers snapshot
		InstanceOnly: true,
		Mode:         "pull",
		// I don't think that it makes sense an incremental update
		// in our use case.
		Refresh: false,
		// Ignore copy errors for volatile files.
		AllowInconsistent: true,
	}

	entry, _, err := e.LxdClient.GetInstance(containerName)
	if err != nil {
		return err
	}

	if entry.Config != nil {
		// Strip the last_state.power key in all cases
		delete(entry.Config, "volatile.last_state.power")
	}

	op, err := e.LxdClient.CopyInstance(e.LxdClient, *entry, &args)
	if err != nil {
		return err
	}

	// Watch the background operation
	progress := lxd_cli.ProgressRenderer{
		Format: "Copy container: %s",
		Quiet:  false,
	}

	_, err = op.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		return err
	}

	// Wait the copy of the container
	err = lxd_cli.CancelableWait(op, &progress)
	if err != nil {
		progress.Done("")
		return err
	}

	progress.Done("")

	e.Emitter.DebugLog(false,
		fmt.Sprintf("Container %s copy to %s.", containerName, newContainerName))

	return nil
}

func (e *LxdExecutor) DeleteContainer(containerName string) error {

	ephemeral, err := e.IsEphemeralContainer(containerName)
	if err != nil {
		e.Emitter.ErrorLog(false,
			fmt.Sprintf("Error on retrieve info of the container %s", containerName))
		return err
	}

	err = e.DoAction2Container(containerName, "stop")
	if err != nil {
		e.Emitter.ErrorLog(false, "Error on stop container: "+err.Error())
		return err
	}

	if !ephemeral {
		var currOper lxd.Operation
		var err error

		// Delete container (set force true considering that is been stopped before)
		currOper, err = e.LxdClient.DeleteInstance(containerName, true)
		if err != nil {
			e.Emitter.ErrorLog(false, "Error on delete container: "+err.Error())
			return err
		}
		_ = e.WaitOperation(currOper, nil)
	}

	return nil
}

func (e *LxdExecutor) WaitIpOfContainer(containerName string, timeout int64) error {
	filters := []string{
		"name=" + containerName,
	}

	start := time.Now().Unix()
	diff := int64(0)
	withoutIp := true
	for withoutIp && diff < timeout {
		instances, err := e.LxdClient.GetInstancesFull(
			lxd.GetInstancesFullArgs{
				InstanceType: lxd_api.InstanceTypeContainer,
				Filters:      filters,
			},
		)
		if err != nil {
			e.Emitter.ErrorLog(false, "Error on get instances: "+err.Error())
			return err
		}

		if len(instances) == 0 {
			return errors.New("No container found with name " + containerName)
		} else if len(instances) > 1 {
			return errors.New("Found multiple container with name " + containerName)
		}

		c := instances[0]
		for netIface, net := range c.State.Network {
			if net.Type == "loopback" {
				continue
			}
			for _, a := range net.Addresses {
				if a.Scope == "link" || a.Scope == "local" {
					continue
				}

				if a.Family == "inet" {
					if a.Address != "" && a.Netmask != "" {
						e.Emitter.Emits(base.LxdContainerIpAssigned, map[string]interface{}{
							"name":    containerName,
							"iface":   netIface,
							"address": fmt.Sprintf("%s/%s", a.Address, a.Netmask),
						})
						withoutIp = false
						break
					}
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
		diff = time.Now().Unix() - start
	}

	return nil
}
