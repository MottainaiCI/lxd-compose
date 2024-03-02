/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"errors"
	"fmt"
	"regexp"
)

type PurgeOpts struct {
	All         bool
	Fingerprint string
	Matches     []string
	NoAliases   bool
}

func (e *LxdCExecutor) PurgeImages(opts *PurgeOpts) error {

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

				for fingerprint, _ := range matchedFingerprints {
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

func (e *LxdCExecutor) DeleteImageByFingerprint(f string) error {
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
