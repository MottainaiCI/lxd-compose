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
	"github.com/jinzhu/copier"
)

func getHooks(hooks *[]LxdCHook, event string) []LxdCHook {
	return getHooks4Nodes(hooks, event, []string{""})
}

func getHooks4Nodes(hooks *[]LxdCHook, event string, nodes []string) []LxdCHook {
	ans := []LxdCHook{}

	if hooks != nil {
		for _, h := range *hooks {
			if h.Event == event {

				for _, node := range nodes {
					if (node == "" && h.Node != "host") || node == "*" {
						ans = append(ans, h)
						break
					} else {
						if node == h.Node {
							ans = append(ans, h)
							break
						}
					}
				}

			}
		}
	}

	return ans
}

func (h *LxdCHook) For(node string) bool {
	if h.Node == "" || h.Node == "*" || h.Node == node {
		return true
	}
	return false
}

func (h *LxdCHook) Clone() *LxdCHook {
	ans := LxdCHook{}
	copier.Copy(&ans, h)
	return &ans
}

func (h *LxdCHook) SetNode(node string) {
	h.Node = node
}

func (h *LxdCHook) ToProcess(enabledFlags, disabledFlags []string) bool {
	ans := false

	if len(h.Flags) == 0 && len(enabledFlags) == 0 {
		return true
	}

	if len(disabledFlags) > 0 {
		// Check if the flag is present
		for _, df := range disabledFlags {
			if h.ContainsFlag(df) {
				return false
			}
		}
	}

	if len(enabledFlags) > 0 {
		for _, ef := range enabledFlags {
			if h.ContainsFlag(ef) {
				ans = true
				break
			}
		}
	} else {
		ans = true
	}

	return ans
}

func (h *LxdCHook) ContainsFlag(flag string) bool {
	ans := false
	if len(h.Flags) > 0 {
		for _, f := range h.Flags {
			if f == flag {
				ans = true
				break
			}
		}
	}

	return ans
}

func FilterHooks4Node(hooks *[]LxdCHook, nodes []string) []LxdCHook {
	ans := []LxdCHook{}

	if hooks != nil {
		for _, h := range *hooks {
			for _, node := range nodes {
				if h.For(node) {
					nh := h.Clone()
					nh.SetNode(node)
					ans = append(ans, *nh)
					break
				}
			}
		}
	}

	return ans
}
