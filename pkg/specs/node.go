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
package specs

import (
	"fmt"
	"path/filepath"

	"github.com/ghodss/yaml"
)

func (n *LxdCNode) Init() {
	if n.Hooks == nil {
		n.Hooks = []LxdCHook{}
	}
}

func (n *LxdCNode) IsSourcePathRelative() bool {
	if filepath.IsAbs(n.SourceDir) {
		return false
	}
	return true
}

func (n *LxdCNode) GetHooks(event string) []LxdCHook {
	return getHooks(&n.Hooks, event)
}

func (n *LxdCNode) GetAllHooks(event string) []LxdCHook {
	return getHooks4Nodes(&n.Hooks, event, []string{"*"})
}

func (n *LxdCNode) ToJson() (string, error) {
	y, err := yaml.Marshal(n)
	if err != nil {
		return "", fmt.Errorf("Error on convert node %s to yaml: %s",
			n.Name, err.Error())
	}

	data, err := yaml.YAMLToJSON(y)
	if err != nil {
		return "", fmt.Errorf("Error on convert node %s to json: %s",
			n.Name, err.Error())
	}

	return string(data), nil
}

func (n *LxdCNode) GetName() string {
	// Note: "-" it's used to avoid override of the name prefix
	if n.NamePrefix != "" && n.NamePrefix != "-" {
		return fmt.Sprintf("%s-%s", n.NamePrefix, n.Name)
	}
	return n.Name
}
