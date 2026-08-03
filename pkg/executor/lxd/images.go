/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package lxd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/MottainaiCI/lxd-compose/pkg/executor/base"
	lxd "github.com/canonical/lxd/client"
)

func (e *LxdExecutor) PurgeImages(opts *base.PurgeOpts) error {

	// NOTE: For now avoid to use GetImagesWithFilter, this
	//       will work yet api_filtering extension is not present
	//       on server and keep control from client.

	images, err := e.LxdClient.GetImages()
	if err != nil {
		return err
	}

	inErr := false
	if len(images) > 0 {
		if opts.All {
			for _, img := range images {
				err = e.DeleteImageByFingerprint(img.Fingerprint)
				if err != nil {
					inErr = true
				}
			}

		} else {

			if opts.NoAliases {

				for _, img := range images {
					if len(img.Aliases) == 0 {
						err = e.DeleteImageByFingerprint(img.Fingerprint)
						if err != nil {
							inErr = true
						}
					}
				}

			}

			if opts.Fingerprint != "" {
				err = e.DeleteImageByFingerprint(opts.Fingerprint)
				if err != nil {
					inErr = true
				}
			}

			if len(opts.Matches) > 0 {
				matchedFingerprints := make(map[string]bool, 0)

				regexes := []*regexp.Regexp{}
				for _, m := range opts.Matches {
					regexes = append(regexes, regexp.MustCompile(m))
				}

				for _, img := range images {
					for idx := range regexes {
						if regexes[idx] == nil {
							continue
						}

						for _, alias := range img.Aliases {
							if regexes[idx].MatchString(alias.Name) {
								matchedFingerprints[img.Fingerprint] = true
								goto nextImg
							}
						}
					}
				nextImg:
				}

				for fingerprint := range matchedFingerprints {
					err = e.DeleteImageByFingerprint(fingerprint)
					if err != nil {
						inErr = true
					}
				}
			}

		}
	}

	if inErr {
		return errors.New("Error on remove one or more images")
	}

	return nil
}

func (e *LxdExecutor) DeleteImageByFingerprint(f string) error {
	op, err := e.LxdClient.DeleteImage(f)
	if err != nil {
		e.Emitter.ErrorLog(false,
			fmt.Sprintf("Error on delete image %s: %s", f, err.Error()))
		return err
	}

	err = e.WaitOperation(op, nil)
	if err != nil {
		e.Emitter.ErrorLog(false,
			fmt.Sprintf("Error on delete image %s: %s", f, err.Error()))
		return err
	}

	e.Emitter.InfoLog(false,
		fmt.Sprintf("Image %s deleted correctly.", f))

	return nil
}

func (e *LxdExecutor) PullImage(imageAlias, imageRemoteServer string) (string, error) {
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

	if e.LxdClient == nil {
		return "", fmt.Errorf("Something goes wrong on initialize client.")
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
