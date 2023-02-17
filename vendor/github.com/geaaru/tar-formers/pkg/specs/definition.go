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
	"io/fs"
	"os"
	"regexp"
	"time"
)

type SpecFile struct {
	File string `yaml:"-" json:"-"`

	// Define the list of prefixes of the path to extract or to inject
	MatchPrefix []string `yaml:"match_prefix,omitempty" json:"match_prefix,omitempty"`
	// Define the list of files to ignore/skip.

	IgnoreFiles []string `yaml:"ignore_files,omitempty" json:"ignore_files,omitempty"`

	// Define the list of regexes used to match the paths to ignore.
	IgnoreRegexes []string `yaml:"ignore_regexes,omitempty" json:"ignore_regexes,omitempty"`

	// If the user handler is set. Permit to define the list of the file where
	// is called the user handler function. If this list is empty it calls
	// the callback every times (and TriggeredMatchesPrefix)
	TriggeredFiles         []string `yaml:"triggered_files,omitempty" json:"triggered_files,omitempty"`
	TriggeredMatchesPrefix []string `yaml:"triggered_matches_prefix,omitempty" json:"triggered_matches_prefix,omitempty"`

	Rename []RenameRule `yaml:"rename,omitempty" json:"rename,omitempty"`

	RemapUids   map[string]string `yaml:"remap_uids,omitempty" json:"remap_uids,omitempty"`
	RemapGids   map[string]string `yaml:"remap_gids,omitempty" json:"remap_gids,omitempty"`
	RemapUsers  map[string]string `yaml:"remap_users,omitempty" json:"remap_users,omitempty"`
	RemapGroups map[string]string `yaml:"remap_groups,omitempty" json:"remap_groups,omitempty"`

	SameOwner        bool `yaml:"same_owner,omitempty" json:"same_owner,omitempty"`
	SameChtimes      bool `yaml:"same_chtimes,omitempty" json:"same_chtimes,omitempty"`
	MapEntities      bool `yaml:"map_entities,omitempty" json:"map_entities,omitempty"`
	BrokenLinksFatal bool `yaml:"broken_links_fatal,omitempty" json:"broken_links_fatal,omitempty"`
	EnableMutex      bool `yaml:"enable_mutex,omitempty" json:"enable_mutex,omitempty"`
	OverwritePerms   bool `yaml:"overwrite_perms,omitempty" json:"overwrite_perms,omitempty"`

	mapModifier   map[string]bool  `yaml:"-" json:"-"`
	ignoreRegexes []*regexp.Regexp `yaml:"-" json:"-"`

	// Parallel max open files.
	MaxOpenFiles int64 `yaml:"max_openfiles,omitempty" json:"max_openfiles,omitempty"`
	BufferSize   int   `yaml:"copy_buffer_size,omitempty" json:"copy_buffer_size,omitempty"`

	// Validate extract when the file is been closed.
	Validate bool `yaml:"validate,omitempty" json:"validate,omitempty"`

	// Writer specific section
	Writer *WriterRules `yaml:"writer,omitempty" json:"writer,omitempty"`
}

type WriterRules struct {
	ArchiveDirs  []string `yaml:"dirs,omitempty" json:"dirs,omitempty"`
	ArchiveFiles []string `yaml:"files,omitempty" json:"files,omitempty"`
}

type RenameRule struct {
	Source string `yaml:"source" json:"source"`
	Dest   string `yaml:"dest" json:"dest"`
}

type FileMeta struct {
	Uid   int    // User ID of owner
	Gid   int    // Group ID of owner
	Uname string // User name of owner
	Gname string // Group name of owner

	FileInfo fs.FileInfo // Permission and mode bits

	ModTime    time.Time // Modification time
	AccessTime time.Time // Access time
	ChangeTime time.Time // Change time

	Xattrs     map[string]string // extend attributes
	PAXRecords map[string]string // PAX extend headers records
}

type Link struct {
	Name     string // Contains the path of the link to create (header.Name)
	Linkname string // Contains the path of the path linked to this link (header.Linkname)
	Path     string // Contains the target path merged to the destination path that must be creatd.
	TypeFlag byte
	Mode     os.FileMode
	Meta     FileMeta
}
