/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"io"
	"os"

	"github.com/MottainaiCI/lxd-compose/pkg/executor/base"
	incus "github.com/MottainaiCI/lxd-compose/pkg/executor/incus"
	lxd "github.com/MottainaiCI/lxd-compose/pkg/executor/lxd"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

type LxdCExecutor interface {
	GetType() string
	Setup() error

	CreateContainer(name, fingerprint, imageServer string, profiles []string) error

	CreateContainerWithConfig(name, fingerprint, imageServer string, profiles []string, configMap map[string]string) error

	StopContainer(name string) error
	StartContainer(name string) error
	GetContainerList() ([]string, error)
	IsRunningContainer(name string) (bool, error)
	IsEphemeralContainer(name string) (bool, error)
	IsPresentContainer(name string) (bool, error)
	CopyContainerOnInstance(srcName, dstName string) error
	DeleteContainer(name string) error
	WaitIpOfContainer(name string, timeout int64) error

	GetAclList() ([]string, error)
	IsPresentACL(name string) (bool, error)
	CreateACL(acl *specs.LxdCAcl) error
	UpdateACL(acl *specs.LxdCAcl) error

	RunCommandWithOutput(name, command string, envs map[string]string,
		outBuffer, errBuffer io.WriteCloser, entrypoint []string,
		uid, git *uint32, cwd string) (int, error)
	RunCommand(name, command string, envs map[string]string,
		entrypoint []string, uid, git *uint32, cwd string) (int, error)
	RunCommandWithOutput4Var(name, command, outVar, errVar string,
		envs *map[string]string, eentrypoint []string,
		uid, git *uint32, cwd string) (int, error)

	RunHostCommandWithOutput(command string, envs map[string]string,
		outBuffer, errBuffer io.WriteCloser, entryPoint []string) (int, error)
	RunHostCommand(command string, envs map[string]string,
		entryPoint []string) (int, error)
	RunHostCommandWithOutput4Var(command, outVar, errVar string,
		envs *map[string]string, entryPoint []string) (int, error)

	RecursiveMkdir(name string, dir string, mode *os.FileMode, uid int64, gid int64) error
	RecursivePushFile(name, source, target string) error
	RecursivePullFile(name string, destPath string, localPath string, localAsTarget bool) error
	DeleteContainerDir(name, dir string) error

	// Images
	PurgeImages(opts *base.PurgeOpts) error
	DeleteImageByFingerprint(f string) error
	PullImage(imageAlias, imageRemoteServer string) (string, error)

	// Profiles
	AddProfiles2Instance(name string, profiles []string) error
	RemoveProfilesFromInstance(name string, profiles []string) error
	GetProfilesList() ([]string, error)
	IsPresentProfile(name string) (bool, error)
	CreateProfile(profile specs.LxdCProfile) error
	UpdateProfile(profile specs.LxdCProfile) error

	// Network
	GetNetworkList() ([]string, error)
	IsPresentNetwork(name string) (bool, error)
	CreateNetwork(net specs.LxdCNetwork) error
	UpdateNetwork(net specs.LxdCNetwork) error
	SyncNetworkForwarders(net *specs.LxdCNetwork) error

	// Storage
	GetStorageList() ([]string, error)
	IsPresentStorage(name string) (bool, error)
	CreateStorage(sto specs.LxdCStorage) error
	UpdateStorage(sto specs.LxdCStorage) error

	// Trust
	GetCertificates() ([]*specs.LxdCCertificate, error)
	DeleteCertificate(fingerprint string) error
	CreateCertificate(cert *specs.LxdCCertificate) error
	IsPresentCertificate(name string) (bool, error)

	// Base
	GetEntrypoint() []string
	SetEntrypoint(ep []string)
	GetEmitter() base.LxdCExecutorEmitter
	SetEmitter(emitter base.LxdCExecutorEmitter)
	SetP2PMode(m bool)
	GetP2PMode() bool
	SetLocalDisable(v bool)
	GetLocalDisable() bool
	SetLegacyApi(a bool)
	GetLegacyApi() bool
	AddRemote2Exclude(remote string)
	SetExcludedRemotes(remotes []string)
	GetExcludedRemotes() []string
	IsRemoteExcluded(remote string) bool
}

func NewLxdCExecutor(connType, endpoint, configdir string, entrypoint []string, ephemeral, showCmdsOutput, runtimeCmdsOutput bool) LxdCExecutor {
	return NewLxdCExecutorWithEmitter(
		connType, endpoint, configdir, entrypoint, ephemeral,
		showCmdsOutput, runtimeCmdsOutput, base.NewLxdCEmitter(),
	)
}

func NewLxdCExecutorWithEmitter(connType, endpoint, configdir string,
	entrypoint []string, ephemeral, showCmdsOutput,
	runtimeCmdsOutput bool, emitter base.LxdCExecutorEmitter) LxdCExecutor {

	if connType == specs.ConnectionLxd6 {
		return lxd.NewLxdExecutorWithEmitter(
			endpoint, configdir, entrypoint, ephemeral,
			showCmdsOutput, runtimeCmdsOutput, emitter)
	} else {
		return incus.NewIncusExecutorWithEmitter(
			endpoint, configdir, entrypoint, ephemeral,
			showCmdsOutput, runtimeCmdsOutput, emitter)
	}

}
