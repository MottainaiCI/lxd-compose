/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"errors"
	"os"
	"os/user"
	"path"
	"strings"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
	lxd "github.com/canonical/lxd/client"
	lxd_config "github.com/canonical/lxd/lxc/config"
)

func getLxcDefaultConfDir() (string, error) {
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

func (e *LxdCExecutor) Setup() error {
	var client lxd.ContainerServer

	configDir, err := getLxcDefaultConfDir()
	if err != nil {
		return errors.New("Error on retrieve default LXD/Incus config directory: " + err.Error())
	}

	if e.ConfigDir == "" {
		e.ConfigDir = configDir
	}
	configPath := path.Join(e.ConfigDir, "/config.yml")
	e.Emitter.DebugLog(false, "Using LXD/Incus config file", configPath)

	e.LxdConfig, err = lxd_config.LoadConfig(configPath)
	if err != nil {
		return errors.New("Error on load LXD/Incus config: " + err.Error())
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

	e.Emitter.Emits(LxdClientSetupDone, map[string]interface{}{
		"defaultRemote": e.LxdConfig.DefaultRemote,
		"configPath":    configPath,
	})

	return nil
}
