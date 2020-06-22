/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ListAll(where string) ([]string, error) {
	var res []string
	files, err := ioutil.ReadDir(where)
	if err != nil {
		return res, err
	}

	for _, file := range files {

		res = append(res, file.Name())

	}
	return res, nil
}

func ListDirs(where string) ([]string, error) {
	var res []string
	files, err := ioutil.ReadDir(where)
	if err != nil {
		return res, err
	}

	for _, file := range files {
		if file.IsDir() {
			res = append(res, file.Name())
		}
	}
	return res, nil
}

// Exists checks if a path exists or not
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// CopyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = CopyFileContents(src, dst)
	return
}

// IsTextFile returns true if file content format is plain text or empty.
func IsTextFile(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return strings.Contains(http.DetectContentType(data), "text/")
}

func IsImageFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "image/")
}

func IsPDFFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "application/pdf")
}

func IsVideoFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "video/")
}

const (
	Byte  = 1
	KByte = Byte * 1024
	MByte = KByte * 1024
	GByte = MByte * 1024
	TByte = GByte * 1024
	PByte = TByte * 1024
	EByte = PByte * 1024
)

var bytesSizeTable = map[string]uint64{
	"b":  Byte,
	"kb": KByte,
	"mb": MByte,
	"gb": GByte,
	"tb": TByte,
	"pb": PByte,
	"eb": EByte,
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+" %s", val, suffix)
}

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}

func TreeList(sourcepath string) []string {
	var artefacts []string
	filepath.Walk(sourcepath, func(path string, f os.FileInfo, err error) error {
		_, file := filepath.Split(path)
		rel := strings.Replace(path, sourcepath, "", 1)
		rel = strings.Replace(rel, file, "", 1)

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			return err
		case mode.IsRegular():
			artefacts = append(artefacts, filepath.Join(rel, file))
		}
		return nil
	})
	return artefacts
}

func DeepCopy(source, destination string) error {
	return filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		_, file := filepath.Split(path)
		rel := strings.Replace(path, source, "", 1)
		rel = strings.Replace(rel, file, "", 1)

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			return err
		case mode.IsRegular():
			os.MkdirAll(filepath.Join(destination, rel), os.ModePerm)
			CopyFile(
				path,
				filepath.Join(destination, rel, file),
			)
		}
		return nil
	})
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
