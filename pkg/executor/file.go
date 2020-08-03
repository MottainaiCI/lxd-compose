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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"

	lxd "github.com/lxc/lxd/client"
	lxd_utils "github.com/lxc/lxd/lxc/utils"
	lxd_shared "github.com/lxc/lxd/shared"
	"github.com/lxc/lxd/shared/ioprogress"
	lxd_units "github.com/lxc/lxd/shared/units"
)

// Based on code of lxc client tool https://github.com/lxc/lxd/blob/master/lxc/file.go
func (e *LxdCExecutor) RecursiveMkdir(nameContainer string, dir string, mode *os.FileMode, uid int64, gid int64) error {

	/* special case, every container has a /, we don't need to do anything */
	if dir == "/" {
		return nil
	}

	// Remove trailing "/" e.g. /A/B/C/. Otherwise we will end up with an
	// empty array entry "" which will confuse the Mkdir() loop below.
	pclean := filepath.Clean(dir)
	parts := strings.Split(pclean, "/")
	i := len(parts)

	for ; i >= 1; i-- {
		cur := filepath.Join(parts[:i]...)
		_, resp, err := e.LxdClient.GetContainerFile(nameContainer, cur)
		if err != nil {
			continue
		}

		if resp.Type != "directory" {
			return fmt.Errorf("%s is not a directory", cur)
		}

		i++
		break
	}

	for ; i <= len(parts); i++ {
		cur := filepath.Join(parts[:i]...)
		if cur == "" {
			continue
		}

		cur = "/" + cur

		modeArg := -1
		if mode != nil {
			modeArg = int(mode.Perm())
		}
		args := lxd.ContainerFileArgs{
			UID:  uid,
			GID:  gid,
			Mode: modeArg,
			Type: "directory",
		}

		log.GetDefaultLogger().Debug(fmt.Sprintf("Creating %s (%s)", cur, args.Type))

		err := e.LxdClient.CreateContainerFile(nameContainer, cur, args)
		if err != nil {
			return err
		}
	}

	return nil
}

// Based on code of lxc client tool https://github.com/lxc/lxd/blob/master/lxc/file.go
func (e *LxdCExecutor) RecursivePushFile(nameContainer, source, target string) error {

	// Determine the target mode
	mode := os.FileMode(0755)
	// Create directory as root. TODO: see if we can use a specific user.
	var uid int64 = 0
	var gid int64 = 0
	err := e.RecursiveMkdir(nameContainer, filepath.Dir(target), &mode, uid, gid)
	if err != nil {
		return err
	}

	//source = filepath.Clean(source)
	//sourceDir, _ := filepath.Split(source)
	sourceDir := filepath.Dir(filepath.Clean(source))
	sourceLen := len(sourceDir)

	sendFile := func(p string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Failed to walk path for %s: %s", p, err)
		}

		// Detect unsupported files
		if !fInfo.Mode().IsRegular() && !fInfo.Mode().IsDir() && fInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			return fmt.Errorf("'%s' isn't a supported file type", p)
		}

		// Prepare for file transfer
		targetPath := path.Join(target, filepath.ToSlash(p[sourceLen:]))
		mode, uid, gid := lxd_shared.GetOwnerMode(fInfo)
		args := lxd.ContainerFileArgs{
			UID:  int64(uid),
			GID:  int64(gid),
			Mode: int(mode.Perm()),
		}

		var readCloser io.ReadCloser

		if fInfo.IsDir() {
			// Directory handling
			args.Type = "directory"
		} else if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			// Symlink handling
			symlinkTarget, err := os.Readlink(p)
			if err != nil {
				return err
			}

			args.Type = "symlink"
			args.Content = bytes.NewReader([]byte(symlinkTarget))
			readCloser = ioutil.NopCloser(args.Content)
		} else {
			// File handling
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()

			args.Type = "file"
			args.Content = f
			readCloser = f
		}

		progress := lxd_utils.ProgressRenderer{
			Format: fmt.Sprintf("Pushing %s to %s: %%s", p, targetPath),
			Quiet:  false,
		}

		if args.Type != "directory" {
			contentLength, err := args.Content.Seek(0, io.SeekEnd)
			if err != nil {
				return err
			}

			_, err = args.Content.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}

			args.Content = lxd_shared.NewReadSeeker(&ioprogress.ProgressReader{
				ReadCloser: readCloser,
				Tracker: &ioprogress.ProgressTracker{
					Length: contentLength,
					Handler: func(percent int64, speed int64) {

						log.GetDefaultLogger().Info(fmt.Sprintf("%d%% (%s/s)", percent,
							lxd_units.GetByteSizeString(speed, 2)))

						progress.UpdateProgress(ioprogress.ProgressData{
							Text: fmt.Sprintf("%d%% (%s/s)", percent,
								lxd_units.GetByteSizeString(speed, 2))})
					},
				},
			}, args.Content)
		}

		log.GetDefaultLogger().Info(fmt.Sprintf("Pushing %s to %s (%s)", p, targetPath, args.Type))
		err = e.LxdClient.CreateContainerFile(nameContainer, targetPath, args)
		if err != nil {
			if args.Type != "directory" {
				progress.Done("")
			}
			return err
		}
		if args.Type != "directory" {
			progress.Done("")
		}
		return nil
	}

	return filepath.Walk(source, sendFile)
}
