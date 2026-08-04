/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package base

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
