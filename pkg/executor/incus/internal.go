/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package incus

import (
	"fmt"
	"time"

	base "github.com/MottainaiCI/lxd-compose/pkg/executor/base"

	incus "github.com/lxc/incus/v7/client"
	incus_api "github.com/lxc/incus/v7/shared/api"
	incus_cli "github.com/lxc/incus/v7/shared/cmd"
)

func (e *IncusExecutor) LaunchContainer(name, fingerprint string, profiles []string) error {
	return e.LaunchContainerType(name, fingerprint, profiles, map[string]string{}, e.Ephemeral)
}

func (e *IncusExecutor) LaunchContainerWithConfig(name, fingerprint string, profiles []string, configMap map[string]string) error {
	return e.LaunchContainerType(name, fingerprint, profiles, configMap, e.Ephemeral)
}

func (e *IncusExecutor) LaunchContainerType(name, fingerprint string, profiles []string, configMap map[string]string, ephemeral bool) error {

	var err error
	var image *incus_api.Image
	var remoteOperation incus.RemoteOperation
	var opInfo *incus_api.Operation

	if len(profiles) == 0 {
		profiles = []string{"default"}
	}

	// Note: Avoid to create devece map for root /. We consider to handle this
	//       as profile. Same for different storage.
	devicesMap := map[string]map[string]string{}

	// Retrieve image info
	image, _, err = e.Client.GetImage(fingerprint)
	if err != nil {
		return err
	}

	// Setup container creation request
	req := incus_api.InstancesPost{
		Name: name,
		Type: incus_api.InstanceTypeContainer,
	}
	req.Config = configMap
	req.Devices = devicesMap
	req.Profiles = profiles
	req.Ephemeral = ephemeral

	// Create the container
	remoteOperation, err = e.Client.CreateInstanceFromImage(e.Client, *image, req)
	if err != nil {
		return fmt.Errorf("error on create instance from image: %s", err)
	}

	// Watch the background operation
	progress := incus_cli.ProgressRenderer{
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
		return fmt.Errorf("error on retrieve opInfo for container: %s", err)
	}

	// LXD instance with operation_metadata_entity_url doesn't supply attributes
	// Resources["instances"]. We can check metadata[entity_url]
	if e.Client.HasExtension("operation_metadata_entity_url") {
		_, ok := opInfo.Metadata["entity_url"]
		if !ok {
			return fmt.Errorf("didn't get any affected image, container or snapshot from server")
		}
	} else {
		instances, ok := opInfo.Resources["instances"]
		if !ok || len(instances) == 0 {
			// Try using the older "containers" field
			instances, ok = opInfo.Resources["containers"]
			if !ok || len(instances) == 0 {
				return fmt.Errorf("didn't get any affected image, container or snapshot from server")
			}
		}
	}

	e.Emitter.Emits(base.LxdContainerCreated, map[string]interface{}{
		"name":      name,
		"profiles":  profiles,
		"ephemeral": e.Ephemeral,
	})

	// Start container
	return e.DoAction2Container(name, "start")
}

func (e *IncusExecutor) WaitOperation(rawOp interface{}, p *incus_cli.ProgressRenderer) error {
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
		err = incus_cli.CancelableWait(rawOp, p)
	} else {
		err = incus_cli.CancelableWait(rawOp, nil)
	}

	return err
}

func (e *IncusExecutor) DoAction2Container(name, action string) error {
	var err error
	var ephemeral bool
	var containerStatus string
	var operation incus.Operation

	instance, _, err := e.Client.GetInstance(name)
	if err != nil {
		if action == "stop" {
			e.Emitter.WarnLog(false,
				fmt.Sprintf("Container %s not found. Already stopped nothing to do.", name))
			return nil
		}
		return err
	}
	ephemeral = instance.Ephemeral
	containerStatus = instance.Status

	if action == "start" && containerStatus == "Started" {
		e.Emitter.WarnLog(false,
			fmt.Sprintf("Container %s is already started!", name))
		return nil
	} else if action == "stop" && containerStatus == "Stopped" {

		if ephemeral {
			// POST: the container is stopped and/or in a weird status. I delete it.
			e.Emitter.WarnLog(false,
				fmt.Sprintf("Container %s is already stopped but ephemeral. Forcing delete.", name))

			var currOper incus.Operation

			// Delete container
			currOper, err = e.Client.DeleteInstance(name)

			if err != nil {
				e.Emitter.ErrorLog(false, "Error on delete container: "+err.Error())
				return err
			}
			_ = e.WaitOperation(currOper, nil)

		} else {
			e.Emitter.WarnLog(false,
				fmt.Sprintf("Container %s is already stopped!", name))
		}
		return nil
	}

	req := incus_api.InstanceStatePut{
		Action:   action,
		Timeout:  120,
		Force:    false,
		Stateful: false,
	}

	operation, err = e.Client.UpdateInstanceState(name, req, "")

	if err != nil {
		e.Emitter.ErrorLog(false, "Error on update container state: "+err.Error())
		return err
	}

	progress := incus_cli.ProgressRenderer{
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
		e.Emitter.Emits(base.LxdContainerStarted, map[string]interface{}{
			"name": name,
		})

	} else {
		e.Emitter.Emits(base.LxdContainerStopped, map[string]interface{}{
			"name": name,
		})
	}

	return nil
}

// Retrieve Image from alias or fingerprint to a specific remote.
func (e *IncusExecutor) GetImage(image string, remote incus.ImageServer) (*incus_api.Image, error) {
	var err error
	var img *incus_api.Image
	var aliasEntry *incus_api.ImageAliasesEntry

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
func (e *IncusExecutor) DeleteImageAliases4Alias(imageAlias string, server incus.InstanceServer) error {
	var err error
	var img *incus_api.Image

	img, _ = e.GetImage(imageAlias, server)
	if img != nil {
		err = e.DeleteImageAliases(img, server)
	}

	return err
}

// Delete all local alias defined on input Image to avoid conflict on pull.
func (e *IncusExecutor) DeleteImageAliases(image *incus_api.Image, server incus.InstanceServer) error {
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

func (e *IncusExecutor) CopyImage(imageFingerprint string, remote incus.ImageServer, to incus.InstanceServer) error {
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
	copyArgs := &incus.ImageCopyArgs{
		Public:     true,
		AutoUpdate: false,
	}

	// Ask LXD to copy the image from the remote server
	// CopyImage return an incus.RemoteOperation does not implement incus.Operation
	// (missing Cancel method) so DownloadImage is not s
	remoteOperation, err := to.CopyImage(remote, *i, copyArgs)
	if err != nil {
		e.Emitter.ErrorLog(false,
			"Error on create copy image task "+err.Error())
		return err
	}

	// Watch the background operation
	progress := incus_cli.ProgressRenderer{
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
		e.AddAlias2Image(i.Fingerprint, alias, e.Client)
	}

	e.Emitter.DebugLog(false, fmt.Sprintf("Image %s copy locally.", imageFingerprint))

	return nil
}

func (e *IncusExecutor) DownloadImage(imageFingerprint string, remote incus.ImageServer) error {
	return e.CopyImage(imageFingerprint, remote, e.Client)
}

func (e *IncusExecutor) AddAlias2Image(fingerprint string, alias incus_api.ImageAlias,
	server incus.InstanceServer) error {
	aliasPost := incus_api.ImageAliasesPost{}
	aliasPost.Name = alias.Name
	aliasPost.Description = alias.Description
	aliasPost.Target = fingerprint
	return server.CreateImageAlias(aliasPost)
}

func (e *IncusExecutor) FindImage(image, imageRemoteServer string) (string, incus.ImageServer, string, error) {
	var err error
	var tmp_srv, srv incus.ImageServer
	var img, tmp_img *incus_api.Image
	var fingerprint string = ""
	var srv_name string = ""

	if imageRemoteServer == "" && !e.P2PMode {
		// Force images if p2m mode is disabled.
		imageRemoteServer = "images"
	}

	for remote, server := range e.Config.Remotes {

		if remote == "local" && e.LocalDisable {
			continue
		}

		if imageRemoteServer != "" && remote != imageRemoteServer && !e.P2PMode {

			e.Emitter.DebugLog(false, fmt.Sprintf(
				"Skipping remote %s. I will use %s.", remote, imageRemoteServer))
			continue
		}

		if e.IsRemoteExcluded(remote) {
			e.Emitter.DebugLog(false, fmt.Sprintf(
				"Skipping remote %s. Remote excluded.", remote))
			continue
		}

		e.Emitter.DebugLog(false, fmt.Sprintf(
			"Found remote %s. I will search the image %s",
			remote, image))
		tmp_srv, err = e.Config.GetImageServer(remote)
		if err != nil {
			err = nil

			e.Emitter.ErrorLog(false, fmt.Sprintf(
				"Error on retrieve ImageServer for remote %s at addr %s",
				remote, server.Addrs[0],
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

func (l *IncusExecutor) CreateImageFromContainer(containerName string, aliases []string,
	properties map[string]string, compressionAlgorithm string, public bool) (string, error) {

	var err error
	imageAliases := []incus_api.ImageAlias{}
	compression := "none"

	// TODO: Check how enable Expires on image created.

	// Check if there is already a local image with same alias. If yes I drop alias.
	for _, aliasName := range aliases {
		aliasEntry, _, _ := l.Client.GetImageAlias(aliasName)
		if aliasEntry != nil {
			l.Emitter.DebugLog(false, fmt.Sprintf(
				"Found old image %s with alias %s. I drop alias from it.",
				aliasEntry.Target, aliasName))

			err = l.Client.DeleteImageAlias(aliasName)
			if err != nil {
				return "", err
			}
		}

		// Reformat aliases
		alias := incus_api.ImageAlias{}
		alias.Name = aliasName
		imageAliases = append(imageAliases, alias)
	}

	if compressionAlgorithm != "" {
		compression = compressionAlgorithm
	}

	// Create the image
	req := incus_api.ImagesPost{
		Source: &incus_api.ImagesPostSource{
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

	op, err := l.Client.CreateImage(req, nil)
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
	_, _, err = l.Client.GetImage(fingerprint)
	if err != nil {
		return "", err
	}

	l.Emitter.InfoLog(false, fmt.Sprintf(
		"For container %s created image %s. Adding aliases %s to image.",
		containerName, fingerprint, aliases))

	for _, alias := range imageAliases {
		aliasPost := incus_api.ImageAliasesPost{}
		aliasPost.Name = alias.Name
		aliasPost.Target = fingerprint
		err := l.Client.CreateImageAlias(aliasPost)
		if err != nil {
			return "", fmt.Errorf("Failed to create alias %s", alias.Name)
		}
	}

	return fingerprint, nil
}
