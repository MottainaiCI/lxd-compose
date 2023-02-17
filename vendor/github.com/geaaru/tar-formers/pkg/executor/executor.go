/*
Copyright (C) 2021-2023  Daniele Rondina <geaaru@gmail.org>

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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/geaaru/tar-formers/pkg/logger"
	specs "github.com/geaaru/tar-formers/pkg/specs"

	"golang.org/x/sync/semaphore"
)

// Mutex must be global to ensure mutual exclusion
// between different TarFormers running.
var mutex sync.Mutex

// Default Tarformers Robot instance
var optimusPrime *TarFormers = nil

type TarFileOperation struct {
	Rename  bool
	NewName string
	Skip    bool
}

// Function handler to
type TarFileHandlerFunc func(path, dst string,
	header *tar.Header, content io.Reader,
	opts *TarFileOperation, t *TarFormers) error

type TarFileWriterHandlerFunc func(path, newpath string,
	header *tar.Header, tw *tar.Writer,
	opts *TarFileOperation, t *TarFormers) error

type TarFormers struct {
	Config *specs.Config `yaml:"config" json:"config"`
	Logger *log.Logger   `yaml:"-" json:"-"`

	reader      io.Reader          `yaml:"-" json:"-"`
	fileHandler TarFileHandlerFunc `yaml:"-" json:"-"`

	writer            io.Writer                `yaml:"-" json:"-"`
	fileWriterHandler TarFileWriterHandlerFunc `yaml:"-" json:"-"`

	Task       *specs.SpecFile `yaml:"task,omitempty" json:"task,omitempty"`
	TaskWriter *specs.SpecFile `yaml:"task_writer,omitempty" json:"task_writer,omitempty"`

	//Using wait group to run f.Sync in parallel
	// Run f.Sync kills time processing.
	waitGroup *sync.WaitGroup
	Ctx       *context.Context
	semaphore *semaphore.Weighted

	flushMutex sync.Mutex
	FlushErrs  []error
}

func SetDefaultTarFormers(t *TarFormers) {
	optimusPrime = t
}

func GetOptimusPrime() *TarFormers {
	return optimusPrime
}

func NewTarFormers(config *specs.Config) *TarFormers {
	return NewTarFormersWithLog(config, false)
}

func NewTarFormersWithLog(config *specs.Config, defLog bool) *TarFormers {
	ans := &TarFormers{
		Config:    config,
		Logger:    log.NewLogger(config),
		Task:      nil,
		waitGroup: &sync.WaitGroup{},
	}

	// Initialize logging
	if config.GetLogging().EnableLogFile && config.GetLogging().Path != "" {
		err := ans.Logger.InitLogger2File()
		if err != nil {
			ans.Logger.Fatal("Error on initialize logfile")
		}
	}

	if defLog {
		ans.Logger.SetAsDefault()
	}
	return ans
}

func (t *TarFormers) SetReader(reader io.Reader) {
	t.reader = reader
}

func (t *TarFormers) SetWriter(writer io.Writer) {
	t.writer = writer
}

func (t *TarFormers) HasFileHandler() bool {
	if t.fileHandler != nil {
		return true
	} else {
		return false
	}
}

func (t *TarFormers) HasFileWriterHandler() bool {
	if t.fileWriterHandler != nil {
		return true
	} else {
		return false
	}
}

func (t *TarFormers) SetFileHandler(f TarFileHandlerFunc) {
	t.fileHandler = f
}

func (t *TarFormers) SetFileWriterHandler(f TarFileWriterHandlerFunc) {
	t.fileWriterHandler = f
}

func (t *TarFormers) RunTaskWriter(task *specs.SpecFile) error {
	if task == nil || task.Writer == nil {
		return errors.New("Invalid task")
	}

	if len(task.Writer.ArchiveDirs) == 0 && len(task.Writer.ArchiveFiles) == 0 {

		return errors.New("No archive dirs or files defined on task")
	}

	t.TaskWriter = task
	t.TaskWriter.Prepare()

	tarWriter := tar.NewWriter(t.writer)
	defer tarWriter.Close()

	err := t.HandleTarFlowWriter(tarWriter)
	if err != nil {
		return err
	}

	return nil
}

func (t *TarFormers) RunTaskBridge(in, out *specs.SpecFile) error {
	if in == nil {
		return errors.New("Invalid input task")
	}

	if out == nil || out.Writer == nil {
		return errors.New("Invalid out task")
	}

	t.TaskWriter = out
	t.Task = in

	t.Task.Prepare()
	t.TaskWriter.Prepare()

	tarWriter := tar.NewWriter(t.writer)
	defer tarWriter.Close()

	tarReader := tar.NewReader(t.reader)

	err := t.HandlerTarBridgeFlow(tarReader, tarWriter)
	if err != nil {
		return err
	}

	return nil
}

func (t *TarFormers) HandlerTarBridgeFlow(
	tarReader *tar.Reader, tarWriter *tar.Writer) error {
	var ans error = nil

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			err = nil
			break
		}

		if err != nil {
			ans = err
			break
		}

		name := header.Name

		// Call file handler also for file that could be skipped and permit
		// to notify this to users.
		if t.HasFileHandler() && t.Task.IsFileTriggered(name) {
			opts := TarFileOperation{
				Rename:  false,
				NewName: "",
				Skip:    false,
			}

			err := t.fileHandler(name, "", header, tarReader, &opts, t)
			if err != nil {
				return err
			}

			if opts.Skip {
				t.Logger.Debug(fmt.Sprintf(
					"File %s skipped from reader callback.", header.Name))
				continue
			}

			if opts.Rename {
				name = opts.NewName
				t.Logger.Debug(fmt.Sprintf(
					"File %s renamed in %s from reader callback.",
					header.Name, name))
			}
		}

		if t.Task.IsPath2Skip(name) {
			t.Logger.Debug(fmt.Sprintf("File %s skipped by reader.", name))
			continue
		}

		fnewname := t.TaskWriter.GetRename(name)

		// Call file handler also for file that could be skipped
		// and permit to notify this to users
		if t.HasFileWriterHandler() && t.TaskWriter.IsFileTriggered(name) {
			opts := TarFileOperation{
				Rename:  false,
				NewName: "",
				Skip:    false,
			}

			err := t.fileWriterHandler(name, fnewname, header, tarWriter, &opts, t)
			if err != nil {
				return fmt.Errorf(
					"Error returned from user handler for file %s: %s",
					name, err.Error())
			}

			if opts.Skip {
				t.Logger.Debug(fmt.Sprintf(
					"File %s skipped from writer callback.", name))
				return nil
			}

			if opts.Rename {
				name = opts.NewName
			} else {
				name = fnewname
			}
		} else if name != fnewname {
			name = fnewname
		}

		if t.TaskWriter.IsPath2Skip(name) {
			t.Logger.Debug(fmt.Sprintf("File %s skipped by writer.", name))
			continue
		}

		t.Logger.Debug(fmt.Sprintf("Processing file %s -> %s of type %d",
			header.Name, name, header.Typeflag))

		header.Name = name

		// Write tar header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf(
				"Error on write header for file '%s': %s'",
				name, err.Error())
		}

		switch header.Typeflag {
		case tar.TypeReg, tar.TypeRegA:
			nb, err := io.Copy(tarWriter, tarReader)
			if err != nil {
				return fmt.Errorf(
					"Error on write file %s: %s", name, err.Error())
			}
			if nb != header.Size {
				return fmt.Errorf(
					"For file %s written %s instead of %s bytes.",
					nb, header.Size)
			}
		}

	}

	tarWriter.Flush()

	return ans
}

func (t *TarFormers) HandleTarFlowWriter(tarWriter *tar.Writer) error {
	imap := make(map[inodeResource]string, 0)

	// Write all directories selected
	if len(t.TaskWriter.Writer.ArchiveDirs) > 0 {
		for _, d := range t.TaskWriter.Writer.ArchiveDirs {
			err := t.InjectDir2Writer(tarWriter, d, &imap)
			if err != nil {
				return fmt.Errorf(
					"Error on inject directory %s: %s", d,
					err.Error())
			}
		}

	}

	// Write all files selected
	if len(t.TaskWriter.Writer.ArchiveFiles) > 0 {
		for _, f := range t.TaskWriter.Writer.ArchiveFiles {
			info, err := os.Stat(f)
			if err != nil {
				return fmt.Errorf(
					"Error on stat file %s: %s", f, err.Error())
			}
			err = t.InjectFile2Writer(tarWriter,
				f, t.TaskWriter.GetRename(f), &info, &imap)
			if err != nil {
				return fmt.Errorf(
					"Error on inject file %s: %s", f, err.Error())
			}
		}
	}

	return nil
}

func (t *TarFormers) RunTask(task *specs.SpecFile, dir string) error {
	if task == nil {
		return errors.New("Invalid task")
	}

	if dir == "" {
		return errors.New("Invalid export dir")
	}

	t.Task = task

	_, err := t.CreateDir(dir, 0755)
	if err != nil {
		return err
	}

	// Setup parallel context and semaphore
	context := context.TODO()
	t.Ctx = &context
	if task.MaxOpenFiles <= 0 {
		task.MaxOpenFiles = 10
	}
	if task.BufferSize <= 0 {
		task.BufferSize = 16
	}
	t.FlushErrs = []error{}
	t.semaphore = semaphore.NewWeighted(task.MaxOpenFiles)

	tarReader := tar.NewReader(t.reader)

	defer t.waitGroup.Wait()

	err = t.HandleTarFlow(tarReader, dir)
	if err != nil {
		return err
	}

	if len(t.FlushErrs) > 0 {
		for _, e := range t.FlushErrs {
			t.Logger.Error(e)
		}
		return errors.New("Received errors on flush files")
	}

	return nil
}

func (t *TarFormers) HandleTarFlow(tarReader *tar.Reader, dir string) error {
	var ans error = nil
	links := []specs.Link{}

	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	t.Task.Prepare()

	for {
		header, err := tarReader.Next()
		newDir := false

		if err == io.EOF {
			err = nil
			break
		}

		if err != nil {
			ans = err
			break
		}

		absPath := "/" + header.Name
		targetPath := filepath.Join(dir, header.Name)
		name := header.Name

		// Call file handler also for file that could be skipped and permit
		// to notify this to users.
		if t.HasFileHandler() && t.Task.IsFileTriggered(absPath) {
			opts := TarFileOperation{
				Rename:  false,
				NewName: "",
				Skip:    false,
			}

			err := t.fileHandler(absPath, dir, header, tarReader, &opts, t)
			if err != nil {
				return err
			}

			if opts.Skip {
				t.Logger.Debug(fmt.Sprintf("File %s skipped.", header.Name))
				continue
			}

			if opts.Rename {
				name = opts.NewName
				if strings.HasPrefix(name, "/") {
					absPath = name
				} else {
					absPath = "/" + name
				}

				targetPath = filepath.Join(dir, name)
			}
		}

		if t.Task.IsPath2Skip(absPath) {
			t.Logger.Debug(fmt.Sprintf("File %s skipped.", name))
			continue
		}

		info := header.FileInfo()

		if t.Config.GetLogging().Level == "debug" {
			t.Logger.Debug(fmt.Sprintf(
				"Parsing file %s [%s - %d, %s - %d] %s (%s).",
				name, header.Uname, header.Uid, header.Gname,
				header.Gid, info.Mode(), header.Linkname))
		}

		switch header.Typeflag {
		case tar.TypeDir:
			newDir, err = t.CreateDir(targetPath, info.Mode())
			if err != nil {
				return fmt.Errorf("Error on create directory %s: %s",
					targetPath, err.Error())
			}
		case tar.TypeReg, tar.TypeRegA:
			err = t.CreateFile(dir, name, info.Mode(), tarReader, header)
			if err != nil {
				return err
			}
		case tar.TypeLink:
			t.Logger.Debug(fmt.Sprintf("Path %s is a hardlink to %s.",
				name, header.Linkname))
			links = append(links,
				specs.Link{
					Path:     targetPath,
					Linkname: filepath.Join(dir, header.Linkname),
					Name:     name,
					Mode:     info.Mode(),
					TypeFlag: header.Typeflag,
					Meta:     specs.NewFileMeta(header),
				})
		case tar.TypeSymlink:
			t.Logger.Debug(fmt.Sprintf("Path %s is a symlink to %s.",
				name, header.Linkname))
			links = append(links,
				specs.Link{
					Path:     targetPath,
					Linkname: header.Linkname,
					Name:     name,
					Mode:     info.Mode(),
					TypeFlag: header.Typeflag,
					Meta:     specs.NewFileMeta(header),
				})
		case tar.TypeChar, tar.TypeBlock:
			err := t.CreateBlockCharFifo(targetPath, info.Mode(), header)
			if err != nil {
				return err
			}

		}

		// Set this an option
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeRegA, tar.TypeBlock, tar.TypeFifo:
			meta := specs.NewFileMeta(header)
			if header.Typeflag != tar.TypeDir || newDir || (!newDir && t.Task.OverwritePerms2Dir()) {
				err := t.SetFileProps(targetPath, &meta, false)
				if err != nil {
					return err
				}
			}
		}

	}

	// Create all links
	if len(links) > 0 {
		//links = t.GetOrderedLinks(links)
		for i := range links {
			err := t.CreateLink(links[i])
			if err != nil {
				return err
			}

			// TODO: check if call setProps to links files too.
			err = t.SetFileProps(links[i].Path, &links[i].Meta, true)
			if err != nil {
				return err
			}
		}
	}

	return ans
}
