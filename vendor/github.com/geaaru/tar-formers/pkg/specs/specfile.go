/*
Copyright (C) 2021-2022  Daniele Rondina <geaaru@funtoo.org>

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
package specs

import (
	"archive/tar"
	"io/fs"
	"io/ioutil"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

func NewSpecFile() *SpecFile {
	return &SpecFile{
		MatchPrefix:            []string{},
		IgnoreFiles:            []string{},
		IgnoreRegexes:          []string{},
		TriggeredFiles:         []string{},
		TriggeredMatchesPrefix: []string{},
		Rename:                 []RenameRule{},
		RemapUids:              make(map[string]string, 0),
		RemapGids:              make(map[string]string, 0),
		RemapUsers:             make(map[string]string, 0),
		RemapGroups:            make(map[string]string, 0),

		SameOwner:        true,
		SameChtimes:      false,
		MapEntities:      false,
		BrokenLinksFatal: false,
		EnableMutex:      false,
		OverwritePerms:   false,
		Validate:         false,

		mapModifier:   make(map[string]bool, 0),
		ignoreRegexes: []*regexp.Regexp{},

		Writer: nil,
	}
}

func NewSpecFileFromYaml(data []byte, f string) (*SpecFile, error) {
	ans := &SpecFile{}
	if err := yaml.Unmarshal(data, ans); err != nil {
		return nil, err
	}

	ans.File = f

	return ans, nil
}

func NewSpecFileFromFile(file string) (*SpecFile, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return NewSpecFileFromYaml(data, file)
}

func (s *SpecFile) OverwritePerms2Dir() bool {
	return s.OverwritePerms
}

func (s *SpecFile) IsFileTriggered(path string) bool {
	if len(s.TriggeredFiles) == 0 && len(s.TriggeredMatchesPrefix) == 0 {
		return true
	}

	if len(s.TriggeredFiles) > 0 {
		if _, p := s.mapModifier[path]; p {
			return true
		}
	}

	if len(s.TriggeredMatchesPrefix) > 0 {
		for _, p := range s.TriggeredMatchesPrefix {
			if strings.HasPrefix(path, p) {
				return true
			}
		}
	}

	return false
}

func (s *SpecFile) Prepare() error {
	// Creating map to speedup research
	s.mapModifier = make(map[string]bool, 0)

	if len(s.TriggeredFiles) > 0 {
		for _, f := range s.TriggeredFiles {
			s.mapModifier[f] = true
		}
	}

	if len(s.IgnoreRegexes) > 0 {
		for _, f := range s.IgnoreRegexes {
			r, err := regexp.Compile(f)
			if err != nil {
				return err
			}

			s.ignoreRegexes = append(s.ignoreRegexes, r)
		}
	}

	return nil
}

func (s *SpecFile) IsPath2Skip(resource string) bool {
	ans := false

	if len(s.MatchPrefix) > 0 {
		for _, p := range s.MatchPrefix {
			if strings.HasPrefix(resource, p) {
				ans = true
				break
			}
		}

		ans = !ans
	}

	if len(s.IgnoreFiles) > 0 && !ans {
		for _, f := range s.IgnoreFiles {
			if f == resource {
				ans = true
				break
			}
		}
	}

	if len(s.IgnoreRegexes) > 0 && !ans {
		for _, r := range s.ignoreRegexes {
			if r.MatchString(resource) {
				ans = true
				break
			}
		}
	}

	return ans
}

func (s *SpecFile) GetRename(file string) string {
	if len(s.Rename) > 0 {
		for _, r := range s.Rename {
			if r.Source == file {
				return r.Dest
			}
		}
	}
	return file
}

func NewFileMeta(header *tar.Header) FileMeta {
	ans := FileMeta{}
	if header != nil {
		ans.Uid = header.Uid
		ans.Gid = header.Gid
		ans.Uname = header.Uname
		ans.Gname = header.Gname
		ans.ModTime = header.ModTime
		ans.AccessTime = header.AccessTime
		ans.ChangeTime = header.ChangeTime
		ans.Xattrs = header.Xattrs
		ans.PAXRecords = header.PAXRecords
		ans.FileInfo = header.FileInfo()
	}
	return ans
}

func (m *FileMeta) GetFileMode() fs.FileMode {
	return m.FileInfo.Mode()
}

func NewWriter() *WriterRules {
	return &WriterRules{
		ArchiveDirs:  []string{},
		ArchiveFiles: []string{},
	}
}

func (w *WriterRules) AddDir(dir string) {
	w.ArchiveDirs = append(w.ArchiveDirs, dir)
}

func (w *WriterRules) AddFile(f string) {
	w.ArchiveFiles = append(w.ArchiveFiles, f)
}
