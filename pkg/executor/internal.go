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
	"time"

	lxd "github.com/lxc/lxd/client"
	lxd_utils "github.com/lxc/lxd/lxc/utils"
	lxd_api "github.com/lxc/lxd/shared/api"
	//"github.com/lxc/lxd/shared/ioprogress"
	//lxd_units "github.com/lxc/lxd/shared/units"
)

func (e *LxdCExecutor) LaunchContainer(name, fingerprint string, profiles []string) error {
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
	req.Ephemeral = e.Ephemeral

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

	err = e.waitOperation(remoteOperation, &progress)
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

	// Start container
	return e.DoAction2Container(name, "start")
}

func (e *LxdCExecutor) waitOperation(rawOp interface{}, p *lxd_utils.ProgressRenderer) error {
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
			fmt.Sprintf("Container %s not found. Already stopped nothing to do.", name)
			return nil
		}
		return err
	}

	if action == "start" && container.Status == "Started" {
		fmt.Sprintf("Container %s is already started!", name)
		return nil
	} else if action == "stop" && container.Status == "Stopped" {
		fmt.Sprintf("Container %s is already stopped!", name)
		return nil
	}

	/*
		if l.Config.GetGeneral().Debug {
			// Permit logging with details about profiles and container
			// configurations only in debug mode.
			l.Report(fmt.Sprintf(
				"Trying to execute action %s to container %s: %v",
				action, name, container,
			))
		} else {
			l.Report(fmt.Sprintf(
				"Executing action %s to container %s...",
				action, name,
			))
		}
	*/

	req := lxd_api.ContainerStatePut{
		Action:   action,
		Timeout:  120,
		Force:    false,
		Stateful: false,
	}

	operation, err = e.LxdClient.UpdateContainerState(name, req, "")
	if err != nil {
		fmt.Println("Error on update container state: " + err.Error())
		return err
	}

	progress := lxd_utils.ProgressRenderer{
		Quiet: false,
	}

	_, err = operation.AddHandler(progress.UpdateOp)
	if err != nil {
		fmt.Println("Error on add handler to progress bar: " + err.Error())
		progress.Done("")
		return err
	}

	err = e.waitOperation(operation, &progress)
	progress.Done("")
	if err != nil {
		fmt.Println(fmt.Sprintf("Error on stop container %s: %s", name, err))
		return err
	}

	if action == "start" {
		fmt.Println(fmt.Sprintf("Container %s is started!", name))
	} else {
		fmt.Println(fmt.Sprintf("Container %s is stopped!", name))
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

		// Check if exists an image with input alias
		aliasEntry, _, err = remote.GetImageAlias(image)
		if err != nil {
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
			fmt.Println(fmt.Sprintf(
				"Found old image %s with alias %s. I drop alias from it.",
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
		fmt.Println("Error on create copy image task " + err.Error())
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

	err = e.waitOperation(remoteOperation, &progress)
	progress.Done("")
	if err != nil {
		fmt.Println("Error on copy image " + err.Error())
		return err
	}

	// Add aliases to images
	for _, alias := range i.Aliases {
		// Ignore error for handle parallel fetching.
		e.AddAlias2Image(i.Fingerprint, alias, e.LxdClient)
	}

	fmt.Println(fmt.Sprintf("Image %s copy locally.", imageFingerprint))

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

func (e *LxdCExecutor) PullImage(imageAlias string) (string, error) {
	var err error
	var imageFingerprint, remote_name string
	var remote lxd.ImageServer

	fmt.Println("Searching image: " + imageAlias)

	// Find image hashing id
	imageFingerprint, remote, remote_name, err = e.FindImage(imageAlias)
	if err != nil {
		return "", err
	}

	if imageFingerprint == imageAlias {
		fmt.Println("Use directly fingerprint " + imageAlias)
	} else {
		fmt.Println("For image " + imageAlias + " found fingerprint " + imageFingerprint)
	}

	// Check if image is already present locally else we receive an error.
	image, _, _ := e.LxdClient.GetImage(imageFingerprint)
	if image == nil {
		// NOTE: In concurrency could be happens that different image that
		//       share same aliases generate reset of aliases but
		//       if I work with fingerprint after FindImage I can ignore
		//       aliases.

		// Delete local image with same target aliases to avoid error on pull.
		err = e.DeleteImageAliases4Alias(imageAlias, e.LxdClient)

		// Try to pull image to lxd instance
		fmt.Println(fmt.Sprintf(
			"Try to download image %s from remote %s...",
			imageFingerprint, remote_name,
		))
		err = e.DownloadImage(imageFingerprint, remote)
	} else {
		fmt.Println("Image " + imageFingerprint + " already present.")
	}

	return imageFingerprint, err
}

func (e *LxdCExecutor) CleanUpContainer(containerName string) error {
	var err error

	err = e.DoAction2Container(containerName, "stop")
	if err != nil {
		fmt.Println("Error on stop container: " + err.Error())
		return err
	}

	if !e.Ephemeral {
		// Delete container
		currOper, err := e.LxdClient.DeleteContainer(containerName)
		if err != nil {
			fmt.Println("Error on delete container: " + err.Error())
			return err
		}
		_ = e.waitOperation(currOper, nil)
	}

	return nil
}

func (l *LxdCExecutor) FindImage(image string) (string, lxd.ImageServer, string, error) {
	var err error
	var tmp_srv, srv lxd.ImageServer
	var img, tmp_img *lxd_api.Image
	var fingerprint string = ""
	var srv_name string = ""

	for remote, server := range l.LxdConfig.Remotes {
		tmp_srv, err = l.LxdConfig.GetImageServer(remote)
		if err != nil {
			err = nil
			fmt.Println(fmt.Sprintf(
				"Error on retrieve ImageServer for remote %s at addr %s",
				remote, server.Addr,
			))
			continue
		}
		tmp_img, err = l.GetImage(image, tmp_srv)
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
