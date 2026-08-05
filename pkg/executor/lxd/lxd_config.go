/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package lxd

import (
	"errors"
	"path"
	"strings"

	"github.com/MottainaiCI/lxd-compose/pkg/executor/base"

	lxd "github.com/canonical/lxd/client"
	lxd_config "github.com/canonical/lxd/lxc/config"
)

func (e *LxdExecutor) Setup() error {
	var client lxd.InstanceServer

	configDir, err := e.GetLxcDefaultConfDir()
	if err != nil {
		return errors.New("Error on retrieve default LXD config directory: " + err.Error())
	}

	if e.ConfigDir == "" {
		e.ConfigDir = configDir
	}
	configPath := path.Join(e.ConfigDir, "/config.yml")
	e.Emitter.DebugLog(false, "Using LXD config file", configPath)

	e.LxdConfig, err = lxd_config.LoadConfig(configPath)
	if err != nil {
		return errors.New("Error on load LXD config: " + err.Error())
	}

	if len(e.Endpoint) > 0 {

		e.Emitter.DebugLog(false, "Using endpoint "+e.Endpoint+"...")

		// Unix socket
		if strings.HasPrefix(e.Endpoint, "unix:") {
			client, err = lxd.ConnectLXDUnix(strings.TrimPrefix(strings.TrimPrefix(e.Endpoint, "unix:"), "//"), nil)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

		} else {
			client, err = e.LxdConfig.GetInstanceServer(e.Endpoint)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

			// Force use of local. Is this needed??
			e.LxdConfig.DefaultRemote = e.Endpoint
		}

	} else {
		if len(e.LxdConfig.DefaultRemote) > 0 {
			// POST: If is present default I use default as main ContainerServer
			client, err = e.LxdConfig.GetInstanceServer(e.LxdConfig.DefaultRemote)
		} else {
			if _, has_local := e.LxdConfig.Remotes["local"]; has_local {
				client, err = e.LxdConfig.GetInstanceServer("local")
				// POST: I use local if is present
			} else {
				// POST: I use default socket connection
				client, err = lxd.ConnectLXDUnix("", nil)
			}
			e.LxdConfig.DefaultRemote = "local"
		}

		if err != nil {
			return errors.New("Error on create LXD Connector: " + err.Error())
		}

	}

	if e.LxdConfig.DefaultRemote == "local" && e.LocalDisable {
		return errors.New("Using local default remote when lxd_local_disable is disable.")
	}

	e.LxdClient = client

	e.Emitter.Emits(base.LxdClientSetupDone, map[string]interface{}{
		"defaultRemote": e.LxdConfig.DefaultRemote,
		"configPath":    configPath,
	})

	return nil
}
