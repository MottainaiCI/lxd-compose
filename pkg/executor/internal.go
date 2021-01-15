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
	"fmt"
	"strings"
	"time"

	lxd "github.com/lxc/lxd/client"
	lxd_utils "github.com/lxc/lxd/lxc/utils"
	lxd_api "github.com/lxc/lxd/shared/api"
)

func (e *LxdCExecutor) LaunchContainer(name, fingerprint string, profiles []string) error {
	return e.LaunchContainerType(name, fingerprint, profiles, e.Ephemeral)
}

func (e *LxdCExecutor) LaunchContainerType(name, fingerprint string, profiles []string, ephemeral bool) error {

	var err error
	var image *lxd_api.Image
	var remoteOperation lxd.RemoteOperation
	var opInfo *lxd_api.Operation

	if len(profiles) == 0 {
		profiles = []string{"default"}
	}

	// Note: Avoid to create devece map for root /. We consider to handle this
	//       as profile. Same for different storage.
	devicesMap := map[string]map[string]string{}
	configMap := map[string]string{}

	// Setup container creation request
	req := lxd_api.ContainersPost{
		Name: name,
	}
	req.Config = configMap
	req.Devices = devicesMap
	req.Profiles = profiles
	req.Ephemeral = ephemeral

	// Retrieve image info
	image, _, err = e.LxdClient.GetImage(fingerprint)
	if err != nil {
		return err
	}

	// Create the container
	remoteOperation, err = e.LxdClient.CreateContainerFromImage(e.LxdClient, *image, req)
	if err != nil {
		return err
	}

	// Watch the background operation
	progress := lxd_utils.ProgressRenderer{
		Format: "Retrieving image: %s",
		Quiet:  false,
	}

	_, err = remoteOperation.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		return err
	}

	err = e.WaitOperation(remoteOperation, &progress)
	if err != nil {
		progress.Done("")
		return err
	}
	progress.Done("")

	// Extract the container name
	opInfo, err = remoteOperation.GetTarget()
	if err != nil {
		return err
	}

	containers, ok := opInfo.Resources["containers"]
	if !ok || len(containers) == 0 {
		return fmt.Errorf("didn't get any affected image, container or snapshot from server")
	}

	e.Emitter.Emits(LxdContainerCreated, map[string]interface{}{
		"name":      name,
		"profiles":  profiles,
		"ephemeral": e.Ephemeral,
	})

	// Start container
	return e.DoAction2Container(name, "start")
}

func (e *LxdCExecutor) WaitOperation(rawOp interface{}, p *lxd_utils.ProgressRenderer) error {
	var err error = nil

	// NOTE: currently on ARM we have a weird behavior where the process that waits
	//       for LXD operation often remain blocked. It seems related to a concurrency
	//       problem on initializing Golang channel.
	//       As a workaround, I sleep some seconds before waiting for a response.

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", e.WaitSleep))
	if err == nil {
		time.Sleep(duration)
	}

	// TODO: Verify if could be a valid idea permit to use wait not cancelable.
	// err = op.Wait()

	if p != nil {
		err = lxd_utils.CancelableWait(rawOp, p)
	} else {
		err = lxd_utils.CancelableWait(rawOp, nil)
	}

	return err
}

func (e *LxdCExecutor) DoAction2Container(name, action string) error {
	var err error
	var container *lxd_api.Container
	var operation lxd.Operation

	container, _, err = e.LxdClient.GetContainer(name)
	if err != nil {
		if action == "stop" {
			e.Emitter.WarnLog(false,
				fmt.Sprintf("Container %s not found. Already stopped nothing to do.", name))
			return nil
		}
		return err
	}

	if action == "start" && container.Status == "Started" {
		e.Emitter.WarnLog(false,
			fmt.Sprintf("Container %s is already started!", name))
		return nil
	} else if action == "stop" && container.Status == "Stopped" {
		e.Emitter.WarnLog(false,
			fmt.Sprintf("Container %s is already stopped!", name))
		return nil
	}

	req := lxd_api.ContainerStatePut{
		Action:   action,
		Timeout:  120,
		Force:    false,
		Stateful: false,
	}

	operation, err = e.LxdClient.UpdateContainerState(name, req, "")
	if err != nil {
		e.Emitter.ErrorLog(false, "Error on update container state: "+err.Error())
		return err
	}

	progress := lxd_utils.ProgressRenderer{
		Quiet: false,
	}

	_, err = operation.AddHandler(progress.UpdateOp)
	if err != nil {
		e.Emitter.ErrorLog(false, "Error on add handler to progress bar: "+err.Error())
		progress.Done("")
		return err
	}

	err = e.WaitOperation(operation, &progress)
	progress.Done("")
	if err != nil {
		e.Emitter.ErrorLog(false,
			fmt.Sprintf("Error on stop container %s: %s", name, err))
		return err
	}

	if action == "start" {
		e.Emitter.Emits(LxdContainerStarted, map[string]interface{}{
			"name": name,
		})

	} else {
		e.Emitter.Emits(LxdContainerStopped, map[string]interface{}{
			"name": name,
		})
	}

	return nil
}

// Retrieve Image from alias or fingerprint to a specific remote.
func (e *LxdCExecutor) GetImage(image string, remote lxd.ImageServer) (*lxd_api.Image, error) {
	var err error
	var img *lxd_api.Image
	var aliasEntry *lxd_api.ImageAliasesEntry

	img, _, err = remote.GetImage(image)
	if err != nil {
		// POST: no image found with input fingerprint
		//       Try to search an image as alias.

		connInfo, _ := remote.GetConnectionInfo()
		remoteURL := ""
		if connInfo != nil {
			remoteURL = connInfo.URL
		}

		// Check if exists an image with input alias
		aliasEntry, _, err = remote.GetImageAlias(image)
		if err != nil {
			e.Emitter.DebugLog(false,
				fmt.Sprintf("On search image with alias %s receive from remote '%s': %s",
					image, remoteURL, err.Error()))
			img = nil
		} else {
			// POST: Find image with alias and so I try to retrieve api.Image
			//       object with all information.
			img, _, err = remote.GetImage(aliasEntry.Target)
		}
	}

	return img, err
}

// Delete alias from image of a specific ContainerServer if available
func (e *LxdCExecutor) DeleteImageAliases4Alias(imageAlias string, server lxd.ContainerServer) error {
	var err error
	var img *lxd_api.Image

	img, _ = e.GetImage(imageAlias, server)
	if img != nil {
		err = e.DeleteImageAliases(img, server)
	}

	return err
}

// Delete all local alias defined on input Image to avoid conflict on pull.
func (e *LxdCExecutor) DeleteImageAliases(image *lxd_api.Image, server lxd.ContainerServer) error {
	for _, alias := range image.Aliases {
		// Retrieve image with alias
		aliasEntry, _, _ := server.GetImageAlias(alias.Name)
		if aliasEntry != nil {
			// TODO: See how handle correctly this use case
			e.Emitter.DebugLog(false,
				fmt.Sprintf("Found old image %s with alias %s. I drop alias from it.",
					aliasEntry.Target, alias.Name))

			err := server.DeleteImageAlias(alias.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *LxdCExecutor) CopyImage(imageFingerprint string, remote lxd.ImageServer, to lxd.ContainerServer) error {
	var err error

	// Get the image information
	i, _, err := remote.GetImage(imageFingerprint)
	if err != nil {
		return err
	}

	// NOTE: we can't use copy aliases here because
	//       LXD doesn't handle correctly concurrency copy
	//       of the same image.
	//       I use i.Aliases after that image is been copied.
	copyArgs := &lxd.ImageCopyArgs{
		Public:     true,
		AutoUpdate: false,
	}

	// Ask LXD to copy the image from the remote server
	// CopyImage return an lxd.RemoteOperation does not implement lxd.Operation
	// (missing Cancel method) so DownloadImage is not s
	remoteOperation, err := to.CopyImage(remote, *i, copyArgs)
	if err != nil {
		e.Emitter.ErrorLog(false,
			"Error on create copy image task "+err.Error())
		return err
	}

	// Watch the background operation
	progress := lxd_utils.ProgressRenderer{
		Format: "Retrieving image: %s",
		Quiet:  false,
	}

	_, err = remoteOperation.AddHandler(progress.UpdateOp)
	if err != nil {
		progress.Done("")
		return err
	}

	err = e.WaitOperation(remoteOperation, &progress)
	progress.Done("")
	if err != nil {
		e.Emitter.ErrorLog(false, "Error on copy image "+err.Error())
		return err
	}

	// Add aliases to images
	for _, alias := range i.Aliases {
		// Ignore error for handle parallel fetching.
		e.AddAlias2Image(i.Fingerprint, alias, e.LxdClient)
	}

	e.Emitter.DebugLog(false, fmt.Sprintf("Image %s copy locally.", imageFingerprint))

	return nil
}

func (e *LxdCExecutor) DownloadImage(imageFingerprint string, remote lxd.ImageServer) error {
	return e.CopyImage(imageFingerprint, remote, e.LxdClient)
}

func (e *LxdCExecutor) AddAlias2Image(fingerprint string, alias lxd_api.ImageAlias,
	server lxd.ContainerServer) error {
	aliasPost := lxd_api.ImageAliasesPost{}
	aliasPost.Name = alias.Name
	aliasPost.Description = alias.Description
	aliasPost.Target = fingerprint
	return server.CreateImageAlias(aliasPost)
}

func (e *LxdCExecutor) PullImage(imageAlias, imageRemoteServer string) (string, error) {
	var err error
	var imageFingerprint, remote_name string
	var remote lxd.ImageServer
	var noRemoteImageFound = false

	e.Emitter.InfoLog(false, "Searching image: "+imageAlias)

	// Find image hashing id
	imageFingerprint, remote, remote_name, err = e.FindImage(imageAlias, imageRemoteServer)
	if err != nil {
		noRemoteImageFound = true
		if strings.Contains(imageAlias, "/") {
			// Is not a fingerprint alias. I can't ensure right image.
			return "", err
		}
		// POST: Try to see if there a local image with the fingerprint
		imageFingerprint = imageAlias
	}

	if imageFingerprint == imageAlias {
		e.Emitter.InfoLog(false, "Use directly fingerprint "+imageAlias)
	} else {
		e.Emitter.InfoLog(false,
			"For image "+imageAlias+" found fingerprint "+imageFingerprint)
	}

	// Check if image is already present locally else we receive an error.
	image, _, _ := e.LxdClient.GetImage(imageFingerprint)
	if image == nil {
		if noRemoteImageFound {
			// No local image found. I return error.
			return "", err
		}

		// NOTE: In concurrency could be happens that different image that
		//       share same aliases generate reset of aliases but
		//       if I work with fingerprint after FindImage I can ignore
		//       aliases.

		// Delete local image with same target aliases to avoid error on pull.
		err = e.DeleteImageAliases4Alias(imageAlias, e.LxdClient)

		// Try to pull image to lxd instance
		e.Emitter.InfoLog(false, fmt.Sprintf(
			"Try to download image %s from remote %s...",
			imageFingerprint, remote_name,
		))
		err = e.DownloadImage(imageFingerprint, remote)
	} else {
		e.Emitter.DebugLog(false,
			"Image "+imageFingerprint+" already present.")
		err = nil
	}

	return imageFingerprint, err
}

func (e *LxdCExecutor) FindImage(image, imageRemoteServer string) (string, lxd.ImageServer, string, error) {
	var err error
	var tmp_srv, srv lxd.ImageServer
	var img, tmp_img *lxd_api.Image
	var fingerprint string = ""
	var srv_name string = ""

	for remote, server := range e.LxdConfig.Remotes {

		if remote == "local" && e.LocalDisable {
			continue
		}

		if imageRemoteServer != "" && remote != imageRemoteServer && !e.P2PMode {

			e.Emitter.DebugLog(false, fmt.Sprintf(
				"Skipping remote %s. I will use %s.", remote, imageRemoteServer))
			continue
		}

		e.Emitter.DebugLog(false, fmt.Sprintf(
			"Found remote %s. I will search the image %s",
			remote, image))
		tmp_srv, err = e.LxdConfig.GetImageServer(remote)
		if err != nil {
			err = nil

			e.Emitter.ErrorLog(false, fmt.Sprintf(
				"Error on retrieve ImageServer for remote %s at addr %s",
				remote, server.Addr,
			))
			continue
		}
		tmp_img, err = e.GetImage(image, tmp_srv)
		if err != nil {
			// POST: No image found with input alias/fingerprint.
			//       I go ahead to next remote
			err = nil
			continue
		}

		if img != nil {
			// POST: A previous image is already found
			if tmp_img.CreatedAt.After(img.CreatedAt) {
				img = tmp_img
				srv = tmp_srv
				srv_name = remote
				fingerprint = img.Fingerprint
			}
		} else {
			// POST: first image matched
			img = tmp_img
			fingerprint = img.Fingerprint
			srv = tmp_srv
			srv_name = remote
		}
	}

	if fingerprint == "" {
		err = fmt.Errorf("No image found with alias or fingerprint %s", image)
	}

	return fingerprint, srv, srv_name, err
}

func (l *LxdCExecutor) CreateImageFromContainer(containerName string, aliases []string, properties map[string]string, compressionAlgorithm string, public bool) (string, error) {

	var err error
	imageAliases := []lxd_api.ImageAlias{}
	compression := "none"

	// TODO: Check how enable Expires on image created.

	// Check if there is already a local image with same alias. If yes I drop alias.
	for _, aliasName := range aliases {
		aliasEntry, _, _ := l.LxdClient.GetImageAlias(aliasName)
		if aliasEntry != nil {
			l.Emitter.DebugLog(false, fmt.Sprintf(
				"Found old image %s with alias %s. I drop alias from it.",
				aliasEntry.Target, aliasName))

			err = l.LxdClient.DeleteImageAlias(aliasName)
			if err != nil {
				return "", err
			}
		}

		// Reformat aliases
		alias := lxd_api.ImageAlias{}
		alias.Name = aliasName
		imageAliases = append(imageAliases, alias)
	}

	if compressionAlgorithm != "" {
		compression = compressionAlgorithm
	}

	// Create the image
	req := lxd_api.ImagesPost{
		Source: &lxd_api.ImagesPostSource{
			Type: "container",
			Name: containerName,
		},
		// CompressionAlgorithm contains name of the binary called by LXD for compression.
		// For any customization create custom script that wrap compression tools.
		CompressionAlgorithm: compression,
	}
	req.Properties = properties
	req.Public = public

	// TODO: Take time and calculate how much time is required for create image
	l.Emitter.InfoLog(false,
		fmt.Sprintf("Starting creation of Image with aliases %s...", aliases))

	op, err := l.LxdClient.CreateImage(req, nil)
	if err != nil {
		return "", err
	}

	err = l.WaitOperation(nil, nil)
	if err != nil {
		return "", err
	}

	opAPI := op.Get()

	// Grab the fingerprint
	fingerprint := opAPI.Metadata["fingerprint"].(string)

	// Get the source image
	_, _, err = l.LxdClient.GetImage(fingerprint)
	if err != nil {
		return "", err
	}

	l.Emitter.InfoLog(false, fmt.Sprintf(
		"For container %s created image %s. Adding aliases %s to image.",
		containerName, fingerprint, aliases))

	for _, alias := range imageAliases {
		aliasPost := lxd_api.ImageAliasesPost{}
		aliasPost.Name = alias.Name
		aliasPost.Target = fingerprint
		err := l.LxdClient.CreateImageAlias(aliasPost)
		if err != nil {
			return "", fmt.Errorf("Failed to create alias %s", alias.Name)
		}
	}

	return fingerprint, nil
}
