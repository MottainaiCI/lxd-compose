/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"errors"
	"fmt"
	"time"

	base "github.com/MottainaiCI/lxd-compose/pkg/executor/base"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	"github.com/MottainaiCI/lxd-compose/pkg/specs"

	incus "github.com/lxc/incus/v7/client"
	incus_api "github.com/lxc/incus/v7/shared/api"
	incus_config "github.com/lxc/incus/v7/shared/cliconfig"
	incus_cli "github.com/lxc/incus/v7/shared/cmd"
)

type IncusExecutor struct {
	*base.BaseExecutor

	Client incus.InstanceServer
	Config *incus_config.Config
}

func NewIncusExecutorWithEmitter(endpoint, configdir string,
	entrypoint []string, ephemeral, showCmdsOutput,
	runtimeCmdsOutput bool, emitter base.LxdCExecutorEmitter) *IncusExecutor {
	return &IncusExecutor{
		BaseExecutor: base.NewBaseExecutorWithEmitter(
			endpoint, configdir, entrypoint,
			ephemeral, showCmdsOutput,
			runtimeCmdsOutput, emitter),
	}
}

func (e *IncusExecutor) GetType() string { return specs.ConnectionIncus }

func (e *IncusExecutor) CreateContainer(name, fingerprint, imageServer string, profiles []string) error {
	return e.CreateContainerWithConfig(name, fingerprint, imageServer, profiles, map[string]string{})
}

func (e *IncusExecutor) CreateContainerWithConfig(name, fingerprint, imageServer string, profiles []string, configMap map[string]string) error {
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

func (e *IncusExecutor) StopContainer(name string) error {
	return e.DoAction2Container(name, "stop")
}

func (e *IncusExecutor) StartContainer(name string) error {
	return e.DoAction2Container(name, "start")
}

func (e *IncusExecutor) GetContainerList() ([]string, error) {
	return e.Client.GetInstanceNames(incus_api.InstanceTypeContainer)
}

func (e *IncusExecutor) IsRunningContainer(name string) (bool, error) {
	ans := false
	var status string

	iInfo, _, err := e.Client.GetInstance(name)
	if err != nil {
		return ans, err
	}
	status = iInfo.Status

	if status == "Running" {
		ans = true
	}

	return ans, nil
}

func (e *IncusExecutor) IsEphemeralContainer(containerName string) (bool, error) {
	iInfo, _, err := e.Client.GetInstance(containerName)
	if err != nil {
		return false, err
	}
	return iInfo.Ephemeral, nil
}

func (e *IncusExecutor) IsPresentContainer(containerName string) (bool, error) {
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

func (e *IncusExecutor) CopyContainerOnInstance(
	containerName, newContainerName string) error {

	args := incus.InstanceCopyArgs{
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

	entry, _, err := e.Client.GetInstance(containerName)
	if err != nil {
		return err
	}

	if entry.Config != nil {
		// Strip the last_state.power key in all cases
		delete(entry.Config, "volatile.last_state.power")
	}

	op, err := e.Client.CopyInstance(e.Client, *entry, &args)
	if err != nil {
		return err
	}

	// Watch the background operation
	progress := incus_cli.ProgressRenderer{
		Format: "Copy container: %s",
		Quiet:  false,
	}

	_, err = op.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		return err
	}

	// Wait the copy of the container
	err = incus_cli.CancelableWait(op, &progress)
	if err != nil {
		progress.Done("")
		return err
	}

	progress.Done("")

	e.Emitter.DebugLog(false,
		fmt.Sprintf("Container %s copy to %s.", containerName, newContainerName))

	return nil
}

func (e *IncusExecutor) DeleteContainer(containerName string) error {

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
		var currOper incus.Operation
		var err error

		// Delete container (set force true considering that is been stopped before)
		currOper, err = e.Client.DeleteInstance(containerName)
		if err != nil {
			e.Emitter.ErrorLog(false, "Error on delete container: "+err.Error())
			return err
		}
		_ = e.WaitOperation(currOper, nil)
	}

	return nil
}

func (e *IncusExecutor) WaitIpOfContainer(containerName string, timeout int64) error {
	filters := []string{
		"name=" + containerName,
	}

	start := time.Now().Unix()
	diff := int64(0)
	withoutIp := true
	for withoutIp && diff < timeout {
		instances, err := e.Client.GetInstancesFullWithFilter(
			incus_api.InstanceTypeContainer,
			filters,
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
