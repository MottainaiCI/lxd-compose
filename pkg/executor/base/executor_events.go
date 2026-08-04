/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package base

type LxdCExecutorEvent string

const (
	LxdClientSetupDone     LxdCExecutorEvent = "client-setup"
	LxdContainerCreated    LxdCExecutorEvent = "container-created"
	LxdContainerUpdated    LxdCExecutorEvent = "container-updated"
	LxdContainerStarted    LxdCExecutorEvent = "container-started"
	LxdContainerStopped    LxdCExecutorEvent = "container-stopped"
	LxdContainerIpAssigned LxdCExecutorEvent = "container-ip"
)

type LxdCExecutorEmitter interface {
	Emits(eType LxdCExecutorEvent, data map[string]interface{})

	DebugLog(color bool, args ...interface{})
	InfoLog(color bool, args ...interface{})
	WarnLog(color bool, args ...interface{})
	ErrorLog(color bool, args ...interface{})
}
