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
	"os"
	"os/exec"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
)

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

	e.Emitter.DebugLog(true,
		logger.Aurora.Bold(
			logger.Aurora.BrightYellow(
				fmt.Sprintf("   :house_with_garden: - entrypoint: %s", entrypoint))))
	e.Emitter.InfoLog(true,
		logger.Aurora.Bold(
			logger.Aurora.BrightYellow("   :house_with_garden: - "+command)))

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
		e.Emitter.InfoLog(false,
			logger.Aurora.Bold(
				logger.Aurora.BrightYellow(
					fmt.Sprintf(">>> [stdout]\n%s", outBuffer.String()))))
	}

	if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
		e.Emitter.InfoLog(false,
			logger.Aurora.Bold(
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
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
					logger.Aurora.BrightYellow(
						fmt.Sprintf(">>> [stdout]\n%s", outBuffer.String()))))
		}

		if e.ShowCmdsOutput && len(errBuffer.String()) > 0 {
			e.Emitter.InfoLog(false,
				logger.Aurora.Bold(
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
