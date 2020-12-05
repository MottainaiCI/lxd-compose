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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"golang.org/x/sys/unix"

	lxd "github.com/lxc/lxd/client"
	lxd_config "github.com/lxc/lxd/lxc/config"
	lxd_api "github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/termios"
)

type LxdCExecutor struct {
	LxdClient      lxd.ContainerServer
	LxdConfig      *lxd_config.Config
	ConfigDir      string
	Endpoint       string
	Entrypoint     []string
	Ephemeral      bool
	ShowCmdsOutput bool
	WaitSleep      int
}

func NewLxdCExecutor(endpoint, configdir string, entrypoint []string, ephemeral, showCmdsOutput bool) *LxdCExecutor {
	return &LxdCExecutor{
		ConfigDir:      configdir,
		Endpoint:       endpoint,
		Entrypoint:     entrypoint,
		Ephemeral:      ephemeral,
		ShowCmdsOutput: showCmdsOutput,
		WaitSleep:      1,
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
	log.GetDefaultLogger().Debug("Using LXD config file", configPath)

	e.LxdConfig, err = lxd_config.LoadConfig(configPath)
	if err != nil {
		return errors.New("Error on load LXD config: " + err.Error())
	}

	if len(e.Endpoint) > 0 {

		log.GetDefaultLogger().Debug("Using endpoint " + e.Endpoint + "...")

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
		logger.InfoC(logger.Aurora.Bold(logger.Aurora.BrightCyan(
			">>> Container " + name + " already present. Nothing to do. - :check_mark:")))
		return nil
	}

	// Pull image
	imageFingerprint, err := e.PullImage(fingerprint, imageServer)
	if err != nil {
		logger.Error("Error on pull image " + fingerprint + " from remote " + imageServer)
		return err
	}

	logger.InfoC(logger.Aurora.Bold(logger.Aurora.BrightCyan(
		">>> Creating container " + name + "... - :factory:")))
	err = e.LaunchContainer(name, imageFingerprint, profiles)
	if err != nil {
		logger.Error("Creating container error: " + err.Error())
		return err
	}

	return nil
}

func (e *LxdCExecutor) DeleteContainer(name string) error {
	return e.CleanUpContainer(name)
}

func (e *LxdCExecutor) RunCommandWithOutput(containerName, command string, envs map[string]string, outBuffer, errBuffer io.WriteCloser, entryPoint []string) (int, error) {
	entrypoint := []string{"/bin/bash", "-c"}

	if len(e.Entrypoint) > 0 {
		entrypoint = e.Entrypoint
	}
	if len(entryPoint) > 0 {
		entrypoint = entryPoint
	}

	if outBuffer == nil {
		return 1, errors.New("Invalid outBuffer")
	}
	if errBuffer == nil {
		return 1, errors.New("Invalid errBuffer")
	}

	// I'm only in Linux/Unix system.
	width, height, err := termios.GetSize(unix.Stdout)

	cmdList := append(entrypoint, command)
	// Prepare the command
	req := lxd_api.ContainerExecPost{
		Command:     cmdList,
		WaitForWS:   true,
		Interactive: false,
		Environment: envs,
		Width:       width,
		Height:      height,
	}

	execArgs := lxd.ContainerExecArgs{
		// Disable stdin
		Stdin:   ioutil.NopCloser(bytes.NewReader(nil)),
		Stdout:  outBuffer,
		Stderr:  errBuffer,
		Control: nil,
		//Control:  handler,
		DataDone: make(chan bool),
	}

	logger := log.GetDefaultLogger()

	logger.DebugC(logger.Aurora.Bold(
		logger.Aurora.BrightCyan(
			fmt.Sprintf(">>> [%s] - entrypoint: %s", containerName, entrypoint))))
	logger.InfoC(logger.Aurora.Italic(
		logger.Aurora.BrightCyan(
			fmt.Sprintf(">>> [%s] - %s - :coffee:", containerName, command))))

	// Run the command in the container
	currOper, err := e.LxdClient.ExecContainer(containerName, req, &execArgs)
	if err != nil {
		logger.Error("Error on exec command: " + err.Error())
		return 1, err
	}

	// Wait for the operation to complete
	err = e.waitOperation(currOper, nil)
	if err != nil {
		logger.Error("Error on waiting execution of commands: " + err.Error())
		return 1, err
	}

	opAPI := currOper.Get()

	// Wait for any remaining I/O to be flushed
	<-execArgs.DataDone

	var ans int
	// NOTE: If I stop a running container for interrupt execution
	// waitOperation doesn't return error but an empty map as opAPI.
	// I consider it as an error.
	if val, ok := opAPI.Metadata["return"]; ok {
		ans = int(val.(float64))
		logger.DebugC(
			logger.Aurora.Bold(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] Exiting [%d]", containerName, ans))))

	} else {
		logger.InfoC(
			logger.Aurora.Bold(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] Execution Interrupted (%v)",
						containerName, opAPI.Metadata))))
		ans = 1
	}

	return ans, nil
}

func (e *LxdCExecutor) RunCommand(containerName, command string, envs map[string]string, entryPoint []string) (int, error) {
	var outBuffer, errBuffer bytes.Buffer
	logger := log.GetDefaultLogger()

	res, err := e.RunCommandWithOutput(containerName, command, envs,
		helpers.NewNopCloseWriter(&outBuffer), helpers.NewNopCloseWriter(&errBuffer),
		entryPoint)

	if err == nil {

		if e.ShowCmdsOutput && len(outBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] [stdout]\n%s", containerName, outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightRed(
					fmt.Sprintf(">>> [%s] [stderr]\n%s", containerName, errBuffer.String()))))
		}
	}

	return res, err
}

func (e *LxdCExecutor) RunCommandWithOutput4Var(containerName, command, outVar, errVar string, envs *map[string]string, entryPoint []string) (int, error) {
	var outBuffer, errBuffer bytes.Buffer
	logger := log.GetDefaultLogger()

	res, err := e.RunCommandWithOutput(containerName, command, *envs,
		helpers.NewNopCloseWriter(&outBuffer), helpers.NewNopCloseWriter(&errBuffer),
		entryPoint)

	if err == nil {

		if e.ShowCmdsOutput && len(outBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] [stdout]\n%s", containerName, outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightRed(
					fmt.Sprintf(">>> [%s] [stderr]\n%s", containerName, errBuffer.String()))))
		}

		if outVar != "" {
			(*envs)[outVar] = outBuffer.String()
		}
		if errVar != "" {
			(*envs)[errVar] = errBuffer.String()
		}
	}

	return res, err
}

func (e *LxdCExecutor) RunHostCommandWithOutput(command string, envs map[string]string, outBuffer, errBuffer io.WriteCloser, entryPoint []string) (int, error) {
	ans := 1

	entrypoint := []string{"/bin/bash", "-c"}
	if len(e.Entrypoint) > 0 {
		entrypoint = e.Entrypoint
	}

	if len(entryPoint) > 0 {
		entrypoint = entryPoint
	}

	if outBuffer == nil {
		return 1, errors.New("Invalid outBuffer")
	}
	if errBuffer == nil {
		return 1, errors.New("Invalid errBuffer")
	}

	cmds := append(entrypoint, command)

	hostCommand := exec.Command(cmds[0], cmds[1:]...)

	logger := log.GetDefaultLogger()

	logger.DebugC(logger.Aurora.Bold(
		logger.Aurora.BrightYellow(
			fmt.Sprintf("   :house_with_garden: - entrypoint: %s", entrypoint))))
	logger.InfoC(logger.Aurora.Bold(
		logger.Aurora.BrightYellow("   :house_with_garden: - " + command)))

	// Convert envs to array list
	elist := os.Environ()
	for k, v := range envs {
		elist = append(elist, k+"="+v)
	}

	if e.ConfigDir != "" {
		elist = append(elist, fmt.Sprintf("LXD_CONF=%s", e.ConfigDir))
	}

	hostCommand.Stdout = outBuffer
	hostCommand.Stderr = errBuffer
	hostCommand.Env = elist

	err := hostCommand.Start()
	if err != nil {
		logger.Error("Error on start command: " + err.Error())
		return 1, err
	}

	err = hostCommand.Wait()
	if err != nil {
		logger.Error("Error on waiting command: " + err.Error())
		return 1, err
	}

	ans = hostCommand.ProcessState.ExitCode()
	logger.DebugC(logger.Aurora.Bold(
		logger.Aurora.BrightYellow(
			fmt.Sprintf("   :house_with_garden: Exiting [%d]", ans))))

	return ans, nil
}

func (e *LxdCExecutor) RunHostCommand(command string, envs map[string]string, entryPoint []string) (int, error) {
	var outBuffer, errBuffer bytes.Buffer
	logger := log.GetDefaultLogger()

	res, err := e.RunHostCommandWithOutput(command, envs,
		helpers.NewNopCloseWriter(&outBuffer), helpers.NewNopCloseWriter(&errBuffer),
		entryPoint)

	if e.ShowCmdsOutput && len(outBuffer.String()) > 0 {
		logger.Info(logger.Aurora.Bold(
			logger.Aurora.BrightYellow(
				fmt.Sprintf(">>> [stdout]\n%s", outBuffer.String()))))
	}

	if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
		logger.Info(logger.Aurora.Bold(
			logger.Aurora.BrightRed(
				fmt.Sprintf(">>> [stderr]\n%s", errBuffer.String()))))
	}

	return res, err
}

func (e *LxdCExecutor) RunHostCommandWithOutput4Var(command, outVar, errVar string, envs *map[string]string, entryPoint []string) (int, error) {
	var outBuffer, errBuffer bytes.Buffer
	logger := log.GetDefaultLogger()

	res, err := e.RunHostCommandWithOutput(command, *envs,
		helpers.NewNopCloseWriter(&outBuffer), helpers.NewNopCloseWriter(&errBuffer),
		entryPoint)

	if err == nil {

		if e.ShowCmdsOutput && len(outBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightYellow(
					fmt.Sprintf(">>> [stdout]\n%s", outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
			logger.Info(logger.Aurora.Bold(
				logger.Aurora.BrightRed(
					fmt.Sprintf(">>> [stderr]\n%s", errBuffer.String()))))
		}

		if outVar != "" {
			(*envs)[outVar] = outBuffer.String()
		}
		if errVar != "" {
			(*envs)[errVar] = errBuffer.String()
		}
	}

	return res, err
}

func (e *LxdCExecutor) GetContainerList() ([]string, error) {
	return e.LxdClient.GetContainerNames()
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

func (e *LxdCExecutor) CleanUpContainer(containerName string) error {
	var err error

	err = e.DoAction2Container(containerName, "stop")
	if err != nil {
		log.GetDefaultLogger().Error("Error on stop container: " + err.Error())
		return err
	}

	if !e.Ephemeral {
		// Delete container
		currOper, err := e.LxdClient.DeleteContainer(containerName)
		if err != nil {
			log.GetDefaultLogger().Error("Error on delete container: " + err.Error())
			return err
		}
		_ = e.waitOperation(currOper, nil)
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
