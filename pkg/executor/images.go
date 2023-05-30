/*
Copyright (C) 2020-2023  Daniele Rondina <geaaru@funtoo.org>
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
