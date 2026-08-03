/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package base

import (
	"os"
	"os/user"
	"path"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
)

type BaseExecutor struct {
	ConfigDir         string
	Endpoint          string
	Entrypoint        []string
	Ephemeral         bool
	ShowCmdsOutput    bool
	RuntimeCmdsOutput bool
	P2PMode           bool
	WaitSleep         int
	LocalDisable      bool
	LegacyApi         bool

	ExcludedRemotes []string

	Emitter LxdCExecutorEmitter
}

type PurgeOpts struct {
	All         bool
	Fingerprint string
	Matches     []string
	NoAliases   bool
}

func NewBaseExecutor(endpoint, configdir string, entrypoint []string, ephemeral, showCmdsOutput, runtimeCmdsOutput bool) *BaseExecutor {
	return NewBaseExecutorWithEmitter(
		endpoint, configdir, entrypoint, ephemeral,
		showCmdsOutput, runtimeCmdsOutput, NewLxdCEmitter(),
	)
}

func NewBaseExecutorWithEmitter(endpoint, configdir string,
	entrypoint []string, ephemeral, showCmdsOutput,
	runtimeCmdsOutput bool, emitter LxdCExecutorEmitter) *BaseExecutor {
	return &BaseExecutor{
		ConfigDir:         configdir,
		Endpoint:          endpoint,
		Entrypoint:        entrypoint,
		Ephemeral:         ephemeral,
		ShowCmdsOutput:    showCmdsOutput,
		RuntimeCmdsOutput: runtimeCmdsOutput,
		WaitSleep:         1,
		Emitter:           emitter,
		P2PMode:           false,
		LocalDisable:      false,
		LegacyApi:         false,
		ExcludedRemotes:   []string{},
	}
}

func (e *BaseExecutor) GetEntrypoint() []string { return e.Entrypoint }

func (e *BaseExecutor) SetEntrypoint(ep []string) { e.Entrypoint = ep }

func (e *BaseExecutor) GetEmitter() LxdCExecutorEmitter        { return e.Emitter }
func (e *BaseExecutor) SetEmitter(emitter LxdCExecutorEmitter) { e.Emitter = emitter }
func (e *BaseExecutor) SetP2PMode(m bool)                      { e.P2PMode = m }
func (e *BaseExecutor) GetP2PMode() bool                       { return e.P2PMode }
func (e *BaseExecutor) SetLocalDisable(v bool)                 { e.LocalDisable = v }
func (e *BaseExecutor) GetLocalDisable() bool                  { return e.LocalDisable }
func (e *BaseExecutor) SetLegacyApi(a bool)                    { e.LegacyApi = a }
func (e *BaseExecutor) GetLegacyApi() bool                     { return e.LegacyApi }

func (e *BaseExecutor) AddRemote2Exclude(remote string) {
	isPresent := false
	for _, r := range e.ExcludedRemotes {
		if r == remote {
			isPresent = true
			break
		}
	}

	if !isPresent {
		e.ExcludedRemotes = append(e.ExcludedRemotes, remote)
	}
}

func (e *BaseExecutor) SetExcludedRemotes(remotes []string) {
	e.ExcludedRemotes = remotes
}

func (e *BaseExecutor) GetExcludedRemotes() []string {
	return e.ExcludedRemotes
}

func (e *BaseExecutor) IsRemoteExcluded(remote string) bool {
	if len(e.ExcludedRemotes) == 0 {
		return false
	}

	for _, r := range e.ExcludedRemotes {
		if r == remote {
			return true
		}
	}

	return false
}

func (e *BaseExecutor) GetLxcDefaultConfDir() (string, error) {
	// Code from LXD project
	var configDir string

	if os.Getenv("LXD_CONF") != "" {
		configDir = os.Getenv("LXD_CONF")
	} else if os.Getenv("INCUS_CONF") != "" {
		configDir = os.Getenv("INCUS_CONF")
	} else if os.Getenv("HOME") != "" {
		incusConfigDir := path.Join(
			os.Getenv("HOME"), ".config", "incus")

		if helpers.Exists(incusConfigDir) {
			configDir = incusConfigDir
		} else {
			configDir = path.Join(os.Getenv("HOME"), ".config", "lxc")
		}
	} else {
		user, err := user.Current()
		if err != nil {
			return "", err
		}

		incusConfigDir := path.Join(user.HomeDir, ".config", "incus")

		if helpers.Exists(incusConfigDir) {
			configDir = incusConfigDir
		} else {
			configDir = path.Join(user.HomeDir, ".config", "lxc")
		}
	}

	return configDir, nil
}
