/*
Copyright © 2020-2024 Daniele Rondina <geaaru@gmail.com>
See AUTHORS and LICENSE for the license details and contributors.
*/
package executor

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"

	lxd "github.com/canonical/lxd/client"
	lxd_shared "github.com/canonical/lxd/shared"
	lxd_cli "github.com/canonical/lxd/shared/cmd"
	"github.com/canonical/lxd/shared/ioprogress"
	lxd_units "github.com/canonical/lxd/shared/units"
)

// Based on code of lxc client tool https://github.com/canonical/lxd/blob/master/lxc/file.go
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
		if e.LegacyApi {
			_, resp, err := e.LxdClient.GetContainerFile(nameContainer, cur)
			if err != nil {
				continue
			}

			if resp.Type != "directory" {
				return fmt.Errorf("%s is not a directory", cur)
			}
		} else {
			_, resp, err := e.LxdClient.GetInstanceFile(nameContainer, cur)
			if err != nil {
				continue
			}

			if resp.Type != "directory" {
				return fmt.Errorf("%s is not a directory", cur)
			}
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

		if e.LegacyApi {
			args := lxd.ContainerFileArgs{
				UID:  uid,
				GID:  gid,
				Mode: modeArg,
				Type: "directory",
			}

			e.Emitter.DebugLog(false, fmt.Sprintf("Creating %s (%s)", cur, args.Type))

			err := e.LxdClient.CreateContainerFile(nameContainer, cur, args)
			if err != nil {
				return err
			}
		} else {
			args := lxd.InstanceFileArgs{
				UID:  uid,
				GID:  gid,
				Mode: modeArg,
				Type: "directory",
			}

			e.Emitter.DebugLog(false, fmt.Sprintf("Creating %s (%s)", cur, args.Type))

			err := e.LxdClient.CreateInstanceFile(nameContainer, cur, args)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Based on code of lxc client tool https://github.com/canonical/lxd/blob/master/lxc/file.go
func (e *LxdCExecutor) RecursivePushFile(nameContainer, source, target string) error {
	var targetIsFile bool = true
	var sourceIsFile bool = true

	if strings.HasSuffix(source, "/") {
		sourceIsFile = false
	}

	if strings.HasSuffix(target, "/") {
		targetIsFile = false
	}

	dir := filepath.Dir(target)
	sourceDir := filepath.Dir(filepath.Clean(source))
	if !sourceIsFile && targetIsFile {
		dir = target
		sourceDir = source
	}
	sourceLen := len(sourceDir)

	// Determine the target mode
	mode := os.FileMode(0755)
	// Create directory as root. TODO: see if we can use a specific user.
	var uid int64 = 0
	var gid int64 = 0
	err := e.RecursiveMkdir(nameContainer, dir, &mode, uid, gid)
	if err != nil {
		return errors.New("Error on create dir " + filepath.Dir(target) + ": " + err.Error())
	}

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

		if p == source {
			if targetIsFile && sourceIsFile {
				targetPath = target
			} else if targetIsFile && !sourceIsFile {
				// Nothing to do. The directory is already been created.
				e.Emitter.DebugLog(false, fmt.Sprintf("Skipping dir %s. Already created.", p))
				return nil
			}
		}

		mode, uid, gid := lxd_shared.GetOwnerMode(fInfo)
		args := lxd.ContainerFileArgs{
			UID:  int64(uid),
			GID:  int64(gid),
			Mode: int(mode.Perm()),
		}

		var readCloser io.ReadCloser
		logger := log.GetDefaultLogger()

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
			readCloser = io.NopCloser(args.Content)
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

		progress := lxd_cli.ProgressRenderer{
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

						if logger.Config.GetLogging().PushProgressBar {
							e.Emitter.InfoLog(true,
								logger.Aurora.Italic(
									logger.Aurora.BrightMagenta(
										fmt.Sprintf("%d%% (%s/s)", percent,
											lxd_units.GetByteSizeString(speed, 2)))))
						}

						progress.UpdateProgress(ioprogress.ProgressData{
							Text: fmt.Sprintf("%d%% (%s/s)", percent,
								lxd_units.GetByteSizeString(speed, 2))})
					},
				},
			}, args.Content)
		}

		if logger.Config.GetLogging().PushProgressBar {
			e.Emitter.InfoLog(true,
				logger.Aurora.Italic(
					logger.Aurora.BrightMagenta(
						fmt.Sprintf(">>> [%s] Pushing %s -> %s (%s)",
							nameContainer, p, targetPath, args.Type))))
		}

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

	sendFileInstance := func(p string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Failed to walk path for %s: %s", p, err)
		}

		// Detect unsupported files
		if !fInfo.Mode().IsRegular() && !fInfo.Mode().IsDir() && fInfo.Mode()&os.ModeSymlink != os.ModeSymlink {
			return fmt.Errorf("'%s' isn't a supported file type", p)
		}

		// Prepare for file transfer
		targetPath := path.Join(target, filepath.ToSlash(p[sourceLen:]))

		if p == source {
			if targetIsFile && sourceIsFile {
				targetPath = target
			} else if targetIsFile && !sourceIsFile {
				// Nothing to do. The directory is already been created.
				e.Emitter.DebugLog(false, fmt.Sprintf("Skipping dir %s. Already created.", p))
				return nil
			}
		}

		mode, uid, gid := lxd_shared.GetOwnerMode(fInfo)
		args := lxd.InstanceFileArgs{
			UID:  int64(uid),
			GID:  int64(gid),
			Mode: int(mode.Perm()),
		}

		var readCloser io.ReadCloser
		logger := log.GetDefaultLogger()

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
			readCloser = io.NopCloser(args.Content)
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

		progress := lxd_cli.ProgressRenderer{
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

						if logger.Config.GetLogging().PushProgressBar {
							e.Emitter.InfoLog(true,
								logger.Aurora.Italic(
									logger.Aurora.BrightMagenta(
										fmt.Sprintf("%d%% (%s/s)", percent,
											lxd_units.GetByteSizeString(speed, 2)))))
						}

						progress.UpdateProgress(ioprogress.ProgressData{
							Text: fmt.Sprintf("%d%% (%s/s)", percent,
								lxd_units.GetByteSizeString(speed, 2))})
					},
				},
			}, args.Content)
		}

		if logger.Config.GetLogging().PushProgressBar {
			e.Emitter.InfoLog(true,
				logger.Aurora.Italic(
					logger.Aurora.BrightMagenta(
						fmt.Sprintf(">>> [%s] Pushing %s -> %s (%s)",
							nameContainer, p, targetPath, args.Type))))
		}

		err = e.LxdClient.CreateInstanceFile(nameContainer, targetPath, args)
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

	if e.LegacyApi {
		return filepath.Walk(source, sendFile)
	} else {
		return filepath.Walk(source, sendFileInstance)
	}
}

// Based on code of lxc client tool https://github.com/canonical/lxd/blob/master/lxc/file.go
func (l *LxdCExecutor) RecursivePullFile(nameContainer string, destPath string, localPath string, localAsTarget bool) error {
	var ftype string
	var mode int
	var entries []string
	var buf io.ReadCloser
	var err error

	if l.LegacyApi {
		var resp *lxd.ContainerFileResponse
		buf, resp, err = l.LxdClient.GetContainerFile(nameContainer, destPath)
		if err != nil {
			return err
		}

		ftype = resp.Type
		mode = resp.Mode
		entries = resp.Entries
	} else {
		var resp *lxd.InstanceFileResponse
		buf, resp, err = l.LxdClient.GetInstanceFile(nameContainer, destPath)
		if err != nil {
			return err
		}

		ftype = resp.Type
		mode = resp.Mode
		entries = resp.Entries
	}

	var target string
	// Default logic is to append tree to target directory
	if localAsTarget {
		target = localPath
	} else {
		target = filepath.Join(localPath, filepath.Base(destPath))
	}
	//target := localPath
	l.Emitter.DebugLog(false, fmt.Sprintf("Pulling %s from %s (%s)\n", target, destPath, ftype))

	if ftype == "directory" {
		err := os.MkdirAll(target, os.FileMode(mode))
		if err != nil {
			l.Emitter.InfoLog(false, fmt.Sprintf("directory %s is already present. Nothing to do.\n", target))
		}

		for _, ent := range entries {
			nextP := path.Join(destPath, ent)

			err = l.RecursivePullFile(nameContainer, nextP, target, false)
			if err != nil {
				return err
			}
		}
	} else if ftype == "file" {
		f, err := os.Create(target)
		if err != nil {
			return err
		}

		defer f.Close()

		err = os.Chmod(target, os.FileMode(mode))
		if err != nil {
			return err
		}

		progress := lxd_cli.ProgressRenderer{
			Format: fmt.Sprintf("Pulling %s from %s: %%s", destPath, target),
			Quiet:  false,
		}

		writer := &ioprogress.ProgressWriter{
			WriteCloser: f,
			Tracker: &ioprogress.ProgressTracker{
				Handler: func(bytesReceived int64, speed int64) {

					l.Emitter.DebugLog(false, fmt.Sprintf("%s (%s/s)\n",
						lxd_units.GetByteSizeString(bytesReceived, 2),
						lxd_units.GetByteSizeString(speed, 2)))

					progress.UpdateProgress(ioprogress.ProgressData{
						Text: fmt.Sprintf("%s (%s/s)",
							lxd_units.GetByteSizeString(bytesReceived, 2),
							lxd_units.GetByteSizeString(speed, 2))})
				},
			},
		}

		_, err = io.Copy(writer, buf)
		progress.Done("")
		if err != nil {
			l.Emitter.ErrorLog(false, fmt.Sprintf("Error on pull file %s", target))
			return err
		}

	} else if ftype == "symlink" {
		linkTarget, err := io.ReadAll(buf)
		if err != nil {
			return err
		}

		err = os.Symlink(strings.TrimSpace(string(linkTarget)), target)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown file type '%s'", ftype)
	}

	return nil
}

func (e *LxdCExecutor) recursiveListFile(nameContainer string, targetPath string, list *list.List) error {
	var ftype string
	var entries []string
	var buf io.ReadCloser
	var err error

	if e.LegacyApi {
		var resp *lxd.ContainerFileResponse
		buf, resp, err = e.LxdClient.GetContainerFile(nameContainer, targetPath)
		if err != nil {
			return err
		}
		ftype = resp.Type
		entries = resp.Entries

	} else {
		var resp *lxd.InstanceFileResponse
		buf, resp, err = e.LxdClient.GetInstanceFile(nameContainer, targetPath)
		if err != nil {
			return err
		}
		ftype = resp.Type
		entries = resp.Entries
	}

	if buf != nil {
		// Needed to avoid: dial unix /var/lib/lxd/unix.socket: socket: too many open files
		buf.Close()
	}

	if ftype == "directory" {
		for _, ent := range entries {
			nextP := path.Join(targetPath, ent)
			err = e.recursiveListFile(nameContainer, nextP, list)
			if err != nil {
				return err
			}
		}
		list.PushBack(targetPath)
	} else if ftype == "file" || ftype == "symlink" {
		list.PushFront(targetPath)

	} else {
		e.Emitter.WarnLog(false, "Find unsupported file type "+ftype+". Skipped.")
	}

	return nil
}

func (e *LxdCExecutor) DeleteContainerDir(name, dir string) error {
	var err error
	var list *list.List = list.New()

	// Create list of files/directories to remove. (files are pushed before directories)
	err = e.recursiveListFile(name, dir, list)
	if err != nil {
		return err
	}

	for f := list.Front(); f != nil; f = f.Next() {
		e.Emitter.DebugLog(false, fmt.Sprintf("Removing file %s...", f.Value.(string)))
		if e.LegacyApi {
			err = e.LxdClient.DeleteContainerFile(name, f.Value.(string))
		} else {
			err = e.LxdClient.DeleteInstanceFile(name, f.Value.(string))
		}
		if err != nil {
			e.Emitter.ErrorLog(false, fmt.Sprintf("ERROR: Error on removing %s: %s",
				f.Value, err.Error()))
		}
	}

	return nil
}
