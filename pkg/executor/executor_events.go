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

type LxdCExecutorEvent string

const (
	LxdClientSetupDone  LxdCExecutorEvent = "client-setup"
	LxdContainerCreated LxdCExecutorEvent = "container-created"
	LxdContainerStarted LxdCExecutorEvent = "container-started"
	LxdContainerStopped LxdCExecutorEvent = "container-stopped"
)

type LxdCExecutorEmitter interface {
	Emits(eType LxdCExecutorEvent, data map[string]interface{})

	DebugLog(color bool, args ...interface{})
	InfoLog(color bool, args ...interface{})
	WarnLog(color bool, args ...interface{})
	ErrorLog(color bool, args ...interface{})
}
