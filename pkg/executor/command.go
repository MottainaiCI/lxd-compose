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

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"

	"golang.org/x/sys/unix"

	lxd "github.com/lxc/lxd/client"
	lxd_api "github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/termios"
)

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

	e.Emitter.DebugLog(true, logger.Aurora.Bold(
		logger.Aurora.BrightCyan(
			fmt.Sprintf(">>> [%s] - entrypoint: %s", containerName, entrypoint))))
	e.Emitter.InfoLog(true, logger.Aurora.Italic(
		logger.Aurora.BrightCyan(
			fmt.Sprintf(">>> [%s] - %s - :coffee:", containerName, command))))

	// Run the command in the container
	currOper, err := e.LxdClient.ExecContainer(containerName, req, &execArgs)
	if err != nil {
		logger.Error("Error on exec command: " + err.Error())
		return 1, err
	}

	// Wait for the operation to complete
	err = e.WaitOperation(currOper, nil)
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
		e.Emitter.DebugLog(true,
			logger.Aurora.Bold(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] Exiting [%d]", containerName, ans))))

	} else {
		e.Emitter.InfoLog(true,
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

		if e.ShowCmdsOutput && len(outBuffer.String()) > 0 && !e.RuntimeCmdsOutput {
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
					logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] [stdout]\n%s", containerName, outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 && !e.RuntimeCmdsOutput {
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
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
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
					logger.Aurora.BrightCyan(
						fmt.Sprintf(">>> [%s] [stdout]\n%s", containerName, outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
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
