/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"errors"
	"path"
	"strings"

	"github.com/MottainaiCI/lxd-compose/pkg/executor/base"

	incus "github.com/lxc/incus/v7/client"
	incus_config "github.com/lxc/incus/v7/shared/cliconfig"
)

func (e *IncusExecutor) Setup() error {
	var client incus.InstanceServer

	configDir, err := e.GetLxcDefaultConfDir()
	if err != nil {
		return errors.New("Error on retrieve default Incus config directory: " + err.Error())
	}

	if e.ConfigDir == "" {
		e.ConfigDir = configDir
	}
	configPath := path.Join(e.ConfigDir, "/config.yml")
	e.Emitter.DebugLog(false, "Using Incus config file", configPath)

	e.Config, err = incus_config.LoadConfig(configPath)
	if err != nil {
		return errors.New("Error on load Incus config: " + err.Error())
	}

	if len(e.Endpoint) > 0 {

		e.Emitter.DebugLog(false, "Using endpoint "+e.Endpoint+"...")

		// Unix socket
		if strings.HasPrefix(e.Endpoint, "unix:") {
			client, err = incus.ConnectIncusUnix(strings.TrimPrefix(strings.TrimPrefix(e.Endpoint, "unix:"), "//"), nil)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

		} else {
			client, err = e.Config.GetInstanceServer(e.Endpoint)
			if err != nil {
				return errors.New("Endpoint:" + e.Endpoint + " Error: " + err.Error())
			}

			// Force use of local. Is this needed??
			e.Config.DefaultRemote = e.Endpoint
		}

	} else {
		if len(e.Config.DefaultRemote) > 0 {
			// POST: If is present default I use default as main ContainerServer
			client, err = e.Config.GetInstanceServer(e.Config.DefaultRemote)
		} else {
			if _, has_local := e.Config.Remotes["local"]; has_local {
				client, err = e.Config.GetInstanceServer("local")
				// POST: I use local if is present
			} else {
				// POST: I use default socket connection
				client, err = incus.ConnectIncusUnix("", nil)
			}
			e.Config.DefaultRemote = "local"
		}

		if err != nil {
			return errors.New("Error on create Incus Connector: " + err.Error())
		}

	}

	if e.Config.DefaultRemote == "local" && e.LocalDisable {
		return errors.New("Using local default remote when lxd_local_disable is disable.")
	}

	e.Client = client

	e.Emitter.Emits(base.LxdClientSetupDone, map[string]interface{}{
		"defaultRemote": e.Config.DefaultRemote,
		"configPath":    configPath,
	})

	return nil
}
