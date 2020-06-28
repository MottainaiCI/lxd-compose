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
package template

import (
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

type LxdCTemplateCompiler interface {
	InitVars()
	SetOpts([]string)
	Compile(sourceFile, destFile string) error
	CompileRaw(sourceContent string) (string, error)
	GetEnvBaseDir() string
	SetEnvBaseDir(string)
	GetVars() *map[string]interface{}
}

type DefaultCompiler struct {
	Project    *specs.LxdCProject
	Opts       []string
	Vars       map[string]interface{}
	EnvBaseDir string
}

func (r *DefaultCompiler) InitVars() {
	r.Vars = make(map[string]interface{}, 0)
	for _, evar := range r.Project.Environments {
		for k, v := range evar.EnvVars {
			r.Vars[k] = v
		}
	}
	// Init project variable
	r.Vars["project"] = r.Project
	r.Vars["env_base_dir"] = r.EnvBaseDir
}

func (r *DefaultCompiler) SetOpts(o []string) {
	r.Opts = o
}

func (r *DefaultCompiler) GetEnvBaseDir() string {
	return r.EnvBaseDir
}

func (r *DefaultCompiler) SetEnvBaseDir(dir string) {
	r.EnvBaseDir = dir
}

func (r *DefaultCompiler) GetVars() *map[string]interface{} {
	return &r.Vars
}
