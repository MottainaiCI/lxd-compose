/*
Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type inodeResource struct {
	Dev uint64
	Ino uint64
}

func (t *TarFormers) InjectFile2Writer(tw *tar.Writer,
	file, fnewname string, stat *fs.FileInfo,
	iMap *map[inodeResource]string) error {

	s := *stat
	imap := *iMap

	if fnewname == "" {
		fnewname = file
	}

	header, err := tar.FileInfoHeader(s, "")
	if err != nil {
		return fmt.Errorf("Error on create tar header for file %s: %s",
			file, err.Error())
	}

	// Call file handler also for file that could be skipped
	// and permit to notify this to users
	if t.HasFileWriterHandler() && t.TaskWriter.IsFileTriggered(file) {
		opts := TarFileOperation{
			Rename:  false,
			NewName: "",
			Skip:    false,
		}

		err := t.fileWriterHandler(file, fnewname, header, tw, &opts, t)
		if err != nil {
			return fmt.Errorf(
				"Error returned from user handler for file %s: %s",
				file, err.Error())
		}

		if opts.Skip {
			t.Logger.Debug(fmt.Sprintf("File %s skipped from user.", file))
			return nil
		}

		if opts.Rename {
			fnewname = opts.NewName
		}
	}

	if t.TaskWriter.IsPath2Skip(file) {
		t.Logger.Debug(fmt.Sprintf("File %s skipped.", file))
		return nil
	}

	xattr, err := t.GetXattr(file)
	if err != nil {
		return fmt.Errorf(
			"Error on get xattr for file %s: %s",
			file, err.Error())
	}

	header.Xattrs = xattr

	stat_t := s.Sys().(*syscall.Stat_t)

	// NOTE: hardlinks and symlinks are not detected correctly
	//       by tar.FileInfoHeader.
	if s.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(file)
		if err != nil {
			return fmt.Errorf(
				"Error on read symbolic link for file %s: %s",
				file, err.Error())
		}
		header.Typeflag = tar.TypeSymlink
		header.Linkname = link
		header.Size = 0
	} else {

		// Register file to inode map.
		in := inodeResource{
			Dev: stat_t.Dev,
			Ino: stat_t.Ino,
		}

		// Check if the file is already present
		orig, ok := imap[in]
		if ok {
			header.Typeflag = tar.TypeLink
			header.Linkname = orig
			header.Size = 0
		} else if !s.IsDir() {
			// TODO: check if convert the link on abs path.
			imap[in] = fnewname
		}
	}

	header.Name = fnewname

	if t.TaskWriter.SameChtimes {
		// Note: this works only on Linux/Unix
		header.AccessTime = time.Unix(stat_t.Atim.Unix())
		header.ChangeTime = time.Unix(stat_t.Ctim.Unix())
	}

	header.Uid = int(stat_t.Uid)
	header.Gid = int(stat_t.Gid)

	t.Logger.Debug(fmt.Sprintf("Processing file %s -> %s of type %d",
		file, header.Name, header.Typeflag))

	err = tw.WriteHeader(header)
	if err != nil {
		return fmt.Errorf(
			"Error on write header for file '%s': %s",
			file, err.Error())
	}

	switch header.Typeflag {
	case tar.TypeDir:
		return nil
	case tar.TypeSymlink:
		t.Logger.Debug(fmt.Sprintf("Injecting symlink %s -> %s",
			header.Name, header.Linkname,
		))
		return nil

	case tar.TypeLink:
		t.Logger.Debug(fmt.Sprintf("Injecting hardlink %s -> %s",
			header.Name, header.Linkname,
		))
		return nil

	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf(
			"Error on open file %s: %s",
			file, err.Error())
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	if err != nil {
		return fmt.Errorf("Error on copy data for file %s: %s",
			file, err.Error())
	}

	return nil
}

func (t *TarFormers) InjectDir2Writer(tw *tar.Writer,
	dir string,
	iMap *map[inodeResource]string) error {

	exists, err := t.ExistFile(dir)
	if err != nil {
		return fmt.Errorf("Error on check if dir %s exists: %s",
			dir, err.Error())
	}

	if !exists {
		return fmt.Errorf("Directory %s doesn't exists.", dir)
	}

	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		return t.InjectFile2Writer(tw, path, t.TaskWriter.GetRename(path), &info, iMap)
	})

	return err
}
