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

type LxdCEmitterWriter struct {
	Type string
}

func NewLxdCEmitterWriter(t string) *LxdCEmitterWriter {
	return &LxdCEmitterWriter{Type: t}
}

func (e *LxdCEmitterWriter) Write(p []byte) (int, error) {
	logger := log.GetDefaultLogger()
	switch e.Type {
	case "lxd_stdout":
		logger.Msg("info", false, false,
			logger.Aurora.Bold(
				logger.Aurora.BrightCyan(string(p)),
			),
		)
	case "host_stdout":
		logger.Msg("info", false, false,
			logger.Aurora.Bold(
				logger.Aurora.BrightYellow(string(p)),
			),
		)
	case "host_stderr", "lxd_stderr":
		logger.Msg("info", false, false,
			logger.Aurora.Bold(
				logger.Aurora.BrightRed(string(p)),
			),
		)
	}
	return len(p), nil
}

func (e *LxdCEmitterWriter) Close() error {
	return nil
}
