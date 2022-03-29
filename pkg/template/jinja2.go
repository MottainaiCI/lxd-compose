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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"gopkg.in/yaml.v3"
)

type Jinja2Compiler struct {
	*DefaultCompiler
}

func NewJinja2Compiler(proj *specs.LxdCProject) *Jinja2Compiler {
	return &Jinja2Compiler{
		DefaultCompiler: &DefaultCompiler{
			Project: proj,
		},
	}
}

func (r *Jinja2Compiler) hasNotYamlEnvs() bool {
	ans := false
	if len(r.Project.IncludeEnvFiles) > 0 {
		for _, file := range r.Project.IncludeEnvFiles {
			if filepath.Ext(file) != ".yml" && filepath.Ext(file) != ".yaml" {
				ans = true
				break
			}
		}
	}

	return ans
}

func (r *Jinja2Compiler) Compile(sourceFile, destFile string) error {
	var dataFile string

	// Create temporary directory
	tmpdir, err := ioutil.TempDir("", "lxd-compose-j2")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	envBaseDirAbs, err := filepath.Abs(r.EnvBaseDir)
	if err != nil {
		return err
	}
	hasNotYmlVars := r.hasNotYamlEnvs()

	if hasNotYmlVars {
		// Then i use the first file.
		// TODO: See if this is correct. We leave to j2cli the read this file.
		// Loader ignore it now.
		dataFile = filepath.Join(envBaseDirAbs, r.Project.IncludeEnvFiles[0])

	} else {
		// Create temporary source file from in memory map
		dataFile = filepath.Join(tmpdir, "data.yml")
		d, err := yaml.Marshal(&r.Vars)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(dataFile, d, 0644)
		if err != nil {
			return err
		}

	}

	// j2cli doesn't create automatically the target directory.
	// I create it before it.
	err = os.MkdirAll(filepath.Dir(destFile), os.ModePerm)
	if err != nil {
		return err
	}

	// Command to execute:
	// j2 template.j2 data.yml -o destfile
	args := []string{
		sourceFile, dataFile,
		"-o", destFile,
	}

	if len(r.Opts) > 0 {
		args = append(args, r.Opts...)
	}
	j2Command := exec.Command("j2", args...)

	j2Command.Stdout = os.Stdout
	j2Command.Stderr = os.Stderr

	err = j2Command.Run()
	if err != nil {
		return err
	}

	return nil
}

func (r *Jinja2Compiler) CompileRaw(sourceData string) (string, error) {

	// Create temporary directory
	tmpdir, err := ioutil.TempDir("", "lxd-compose-j2")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpdir)

	sourceFile := filepath.Join(tmpdir, "source.j2")
	destFile := filepath.Join(tmpdir, "dest.yaml")
	err = ioutil.WriteFile(sourceFile, []byte(sourceData), 0644)
	if err != nil {
		return "", err
	}

	err = r.Compile(sourceFile, destFile)
	if err != nil {
		return "", err
	}

	ans, err := ioutil.ReadFile(destFile)
	if err != nil {
		return "", err
	}

	return string(ans), nil
}
