/*
Copyright (C) 2020-2023  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"os"
	"path/filepath"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

type MottainaiCompiler struct {
	*DefaultCompiler
}

func NewMottainaiCompiler(proj *specs.LxdCProject) *MottainaiCompiler {
	return &MottainaiCompiler{
		DefaultCompiler: &DefaultCompiler{
			Project: proj,
		},
	}
}

func (r *MottainaiCompiler) Compile(sourceFile, destFile string) error {

	sourceData, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	dstData, err := r.CompileRaw(string(sourceData))
	if err != nil {
		return err
	}

	dir := filepath.Dir(destFile)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(destFile, []byte(dstData), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *MottainaiCompiler) CompileRaw(sourceData string) (string, error) {
	tmpl := NewTemplate()
	tmpl.Values = r.Vars

	destData, err := tmpl.Draw(sourceData)
	if err != nil {
		return "", err
	}

	return destData, nil
}
