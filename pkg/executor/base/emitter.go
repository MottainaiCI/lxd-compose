/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package base

import (
	"fmt"
	"io"

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
)

type LxdCEmitter struct {
	HostWriterStdout io.WriteCloser
	HostWriterStderr io.WriteCloser
	LxdWriterStdout  io.WriteCloser
	LxdWriterStderr  io.WriteCloser
}

func NewLxdCEmitter() *LxdCEmitter {
	return &LxdCEmitter{
		HostWriterStdout: NewLxdCEmitterWriter("host_stdout"),
		HostWriterStderr: NewLxdCEmitterWriter("host_stderr"),
		LxdWriterStdout:  NewLxdCEmitterWriter("lxd_stdout"),
		LxdWriterStderr:  NewLxdCEmitterWriter("lxd_stderr"),
	}
}

func (e *LxdCEmitter) GetHostWriterStdout() io.WriteCloser  { return e.HostWriterStdout }
func (e *LxdCEmitter) GetHostWriterStderr() io.WriteCloser  { return e.HostWriterStderr }
func (e *LxdCEmitter) SetHostWriterStdout(w io.WriteCloser) { e.HostWriterStdout = w }
func (e *LxdCEmitter) SetHostWriterStderr(w io.WriteCloser) { e.HostWriterStderr = w }

func (e *LxdCEmitter) GetLxdWriterStdout() io.WriteCloser  { return e.LxdWriterStdout }
func (e *LxdCEmitter) GetLxdWriterStderr() io.WriteCloser  { return e.LxdWriterStderr }
func (e *LxdCEmitter) SetLxdWriterStdout(w io.WriteCloser) { e.LxdWriterStdout = w }
func (e *LxdCEmitter) SetLxdWriterStderr(w io.WriteCloser) { e.LxdWriterStderr = w }

func (e *LxdCEmitter) DebugLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("debug", color, true, args...)
}

func (e *LxdCEmitter) InfoLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("info", color, true, args...)
}

func (e *LxdCEmitter) WarnLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("warning", color, true, args...)
}

func (e *LxdCEmitter) ErrorLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("error", color, true, args...)
}

func (e *LxdCEmitter) Emits(eType LxdCExecutorEvent, data map[string]interface{}) {
	logger := log.GetDefaultLogger()

	// TODO: review management of the setup event. We reload config too many times.
	switch eType {
	case LxdContainerStarted:
		e.InfoLog(true,
			logger.Aurora.Bold(logger.Aurora.BrightCyan(
				">>> ["+data["name"].(string)+"] - [started] :bomb:")))

	case LxdContainerStopped:
		e.InfoLog(true,
			logger.Aurora.Bold(logger.Aurora.BrightCyan(
				">>> ["+data["name"].(string)+"] - [stopped] :check_mark:")))
	case LxdContainerIpAssigned:
		e.InfoLog(true,
			logger.Aurora.Bold(logger.Aurora.BrightCyan(
				fmt.Sprintf(
					">>> [%s] - [%s] %s :computer:",
					data["name"].(string),
					data["iface"].(string),
					data["address"].(string),
				))))
	}
}
