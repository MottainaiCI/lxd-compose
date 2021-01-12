/*

Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"os"
	"os/user"
	"path"
	"strings"

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	lxd "github.com/lxc/lxd/client"
	lxd_config "github.com/lxc/lxd/lxc/config"
	lxd_api "github.com/lxc/lxd/shared/api"
)

type LxdCExecutor struct {
	LxdClient         lxd.ContainerServer
	LxdConfig         *lxd_config.Config
	ConfigDir         string
	Endpoint          string
	Entrypoint        []string
	Ephemeral         bool
	ShowCmdsOutput    bool
	RuntimeCmdsOutput bool
	WaitSleep         int

	Emitter LxdCExecutorEmitter
}

func NewLxdCExecutor(endpoint, configdir string, entrypoint []string, ephemeral, showCmdsOutput, runtimeCmdsOutput bool) *LxdCExecutor {
	return NewLxdCExecutorWithEmitter(
		endpoint, configdir, entrypoint, ephemeral,
		showCmdsOutput, runtimeCmdsOutput, NewLxdCEmitter(),
	)
}

func NewLxdCExecutorWithEmitter(endpoint, configdir string,
	entrypoint []string, ephemeral, showCmdsOutput,
	runtimeCmdsOutput bool, emitter LxdCExecutorEmitter) *LxdCExecutor {
	return &LxdCExecutor{
		ConfigDir:         configdir,
		Endpoint:          endpoint,
		Entrypoint:        entrypoint,
		Ephemeral:         ephemeral,
		ShowCmdsOutput:    showCmdsOutput,
		RuntimeCmdsOutput: runtimeCmdsOutput,
		WaitSleep:         1,
		Emitter:           emitter,
	}
}

func getLxcDefaultConfDir() (string, error) {
	// Code from LXD project
	var configDir string

	if os.Getenv("LXD_CONF") != "" {
		configDir = os.Getenv("LXD_CONF")
	} else if os.Getenv("HOME") != "" {
		configDir = path.Join(os.Getenv("HOME"), ".config", "lxc")
	} else {
		user, err := user.Current()
		if err != nil {
			return "", err
		}

		configDir = path.Join(user.HomeDir, ".config", "lxc")
	}

	return configDir, nil
}

func (e *LxdCExecutor) GetEmitter() LxdCExecutorEmitter        { return e.Emitter }
func (e *LxdCExecutor) SetEmitter(emitter LxdCExecutorEmitter) { e.Emitter = emitter }

func (e *LxdCExecutor) Setup() error {
	var client lxd.ContainerServer

	configDir, err := getLxcDefaultConfDir()
	if err != nil {
		return errors.New("Error on retrieve default LXD config directory: " + err.Error())
	}

	if e.ConfigDir == "" {
		e.ConfigDir = configDir
	}
	configPath := path.Join(e.ConfigDir, "/config.yml")
	e.Emitter.DebugLog(false, "Using LXD config file", configPath)

	e.LxdConfig, err = lxd_config.LoadConfig(configPath)
	if err != nil {
		return errors.New("Error on load LXD config: " + err.Error())
	}

	if len(e.Endpoint) > 0 {

		e.Emitter.DebugLog(false, "Using endpoint "+e.Endpoint+"...")

		// Unix socket
		if strings.HasPrefix(e.Endpoint, "unix:") {
			client, err = lxd.ConnectLXDUnix(strings.TrimPrefix(strings.TrimPrefix(e.Endpoint, "unix:"), "//"), nil)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

		} else {
			client, err = e.LxdConfig.GetInstanceServer(e.Endpoint)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

			// Force use of local. Is this needed??
			e.LxdConfig.DefaultRemote = e.Endpoint
		}

	} else {
		if len(e.LxdConfig.DefaultRemote) > 0 {
			// POST: If is present default I use default as main ContainerServer
			client, err = e.LxdConfig.GetInstanceServer(e.LxdConfig.DefaultRemote)
		} else {
			if _, has_local := e.LxdConfig.Remotes["local"]; has_local {
				client, err = e.LxdConfig.GetInstanceServer("local")
				// POST: I use local if is present
			} else {
				// POST: I use default socket connection
				client, err = lxd.ConnectLXDUnix("", nil)
			}
			if err != nil {
				return errors.New("Error on create LXD Connector: " + err.Error())
			}

			e.LxdConfig.DefaultRemote = "local"
		}
	}

	e.LxdClient = client

	e.Emitter.Emits(LxdClientSetupDone, map[string]interface{}{
		"defaultRemote": e.LxdConfig.DefaultRemote,
		"configPath":    configPath,
	})

	return nil
}

func (e *LxdCExecutor) CreateContainer(name, fingerprint, imageServer string, profiles []string) error {
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
	err = e.LaunchContainer(name, imageFingerprint, profiles)
	if err != nil {
		logger.Error("Creating container error: " + err.Error())
		return err
	}

	return nil
}

func (e *LxdCExecutor) StopContainer(name string) error {
	return e.DoAction2Container(name, "stop")
}

func (e *LxdCExecutor) StartContainer(name string) error {
	return e.DoAction2Container(name, "start")
}

func (e *LxdCExecutor) GetContainerList() ([]string, error) {
	return e.LxdClient.GetContainerNames()
}

func (e *LxdCExecutor) IsEphemeralContainer(containerName string) (bool, error) {
	ans := false

	cInfo, _, err := e.LxdClient.GetContainer(containerName)
	if err != nil {
		return ans, err
	}

	return cInfo.ContainerPut.Ephemeral, nil
}

func (e *LxdCExecutor) IsPresentContainer(containerName string) (bool, error) {
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

func (e *LxdCExecutor) DeleteContainer(containerName string) error {

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
		// Delete container
		currOper, err := e.LxdClient.DeleteContainer(containerName)
		if err != nil {
			e.Emitter.ErrorLog(false, "Error on delete container: "+err.Error())
			return err
		}
		_ = e.WaitOperation(currOper, nil)
	}

	return nil
}

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
			Config:  profile.Config,
			Devices: profile.Devices,
		},
		Name: profile.Name,
	}

	if lxdProfile.ProfilePut.Config == nil {
		lxdProfile.ProfilePut.Config = make(map[string]string, 0)
	}
	if lxdProfile.ProfilePut.Devices == nil {
		lxdProfile.ProfilePut.Devices = make(map[string]map[string]string, 0)
	}

	return e.LxdClient.CreateProfile(lxdProfile)
}
