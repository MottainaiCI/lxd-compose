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
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
)

type LxdCEmitter struct{}

func NewLxdCEmitter() *LxdCEmitter { return &LxdCEmitter{} }

func (e *LxdCEmitter) DebugLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("debug", color, args...)
}

func (e *LxdCEmitter) InfoLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("info", color, args...)
}

func (e *LxdCEmitter) WarnLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("warning", color, args...)
}

func (e *LxdCEmitter) ErrorLog(color bool, args ...interface{}) {
	log.GetDefaultLogger().Msg("error", color, args...)
}

func (e *LxdCEmitter) Emits(eType LxdCExecutorEvent, data map[string]interface{}) {
	logger := log.GetDefaultLogger()

	// TODO: review management of the setup event. We reload config too many times.
	switch eType {
	case LxdContainerStarted:
		e.InfoLog(true,
			logger.Aurora.Bold(logger.Aurora.BrightCyan(
				">>> ["+data["name"].(string)+"] - [stopped] :bomb:")))

	case LxdContainerStopped:
		e.InfoLog(true,
			logger.Aurora.Bold(logger.Aurora.BrightCyan(
				">>> ["+data["name"].(string)+"] - [started] :check_mark:")))
	}
}
