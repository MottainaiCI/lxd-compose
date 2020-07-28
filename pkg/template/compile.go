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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

type CompilerOpts struct {
	Sources []string
}

func NewProjectTemplateCompiler(env *specs.LxdCEnvironment, proj *specs.LxdCProject) (LxdCTemplateCompiler, error) {
	var compiler LxdCTemplateCompiler

	switch env.TemplateEngine.Engine {
	case "jinja2":
		compiler = NewJinja2Compiler(proj)
	case "mottainai":
		compiler = NewMottainaiCompiler(proj)
	default:
		return compiler, errors.New("Invalid template engine " + env.TemplateEngine.Engine)
	}

	compiler.SetEnvBaseDir(filepath.Dir(env.File))
	compiler.SetOpts(env.TemplateEngine.Opts)
	compiler.InitVars()

	return compiler, nil
}

func CompileAllProjectFiles(env *specs.LxdCEnvironment, pName string, opts CompilerOpts) error {

	proj := env.GetProjectByName(pName)
	compiler, err := NewProjectTemplateCompiler(env, proj)
	if err != nil {
		return err
	}

	// Compile project files
	err = CompileProjectFiles(proj, compiler, opts)
	if err != nil {
		return err
	}

	// TODO: parallel elaboration
	for _, group := range proj.Groups {

		// Compile group files
		err = CompileGroupFiles(&group, compiler, opts)
		if err != nil {
			return err
		}

		for _, node := range group.Nodes {
			err := CompileNodeFiles(node, compiler, opts)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CompileGroupFiles(group *specs.LxdCGroup, compiler LxdCTemplateCompiler, opts CompilerOpts) error {
	var sourceFile, destFile string
	var targets []specs.LxdCConfigTemplate = []specs.LxdCConfigTemplate{}

	if len(opts.Sources) > 0 {
		for _, s := range opts.Sources {
			for _, ct := range group.ConfigTemplates {
				if strings.HasPrefix(ct.Source, s) {
					targets = append(targets, ct)
					break
				}
			}
		}
	} else {
		targets = group.ConfigTemplates
	}

	// Set node key with current group
	(*compiler.GetVars())["group"] = group

	for _, s := range targets {
		sourceFile = filepath.Join(compiler.GetEnvBaseDir(), s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = compiler.GetEnvBaseDir()
		}

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		fmt.Println(" " + sourceFile + " -> " + destFile + " OK")
	}

	return nil
}

func CompileProjectFiles(proj *specs.LxdCProject, compiler LxdCTemplateCompiler, opts CompilerOpts) error {
	var sourceFile, destFile string
	var targets []specs.LxdCConfigTemplate = []specs.LxdCConfigTemplate{}

	if len(opts.Sources) > 0 {
		for _, s := range opts.Sources {
			for _, ct := range proj.ConfigTemplates {
				if strings.HasPrefix(ct.Source, s) {
					targets = append(targets, ct)
					break
				}
			}
		}
	} else {
		targets = proj.ConfigTemplates
	}

	// Set node key with current proj
	(*compiler.GetVars())["project"] = proj

	for _, s := range targets {
		sourceFile = filepath.Join(compiler.GetEnvBaseDir(), s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = compiler.GetEnvBaseDir()
		}

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		fmt.Println(" " + sourceFile + " -> " + destFile + " OK")
	}

	return nil
}

func CompileNodeFiles(node specs.LxdCNode, compiler LxdCTemplateCompiler, opts CompilerOpts) error {
	var sourceFile, destFile, baseDir string
	var targets []specs.LxdCConfigTemplate = []specs.LxdCConfigTemplate{}

	if len(opts.Sources) > 0 {
		for _, s := range opts.Sources {

			for _, ct := range node.ConfigTemplates {
				if strings.HasPrefix(ct.Source, s) {
					targets = append(targets, ct)
					break
				}
			}
		}
	} else {
		targets = node.ConfigTemplates
	}

	if len(targets) == 0 {
		return nil
	}

	// Set node key with current node
	(*compiler.GetVars())["node"] = node

	if filepath.IsAbs(node.SourceDir) {
		baseDir = node.SourceDir
	} else {
		baseDir = filepath.Join(compiler.GetEnvBaseDir(), node.SourceDir)
	}

	for _, s := range targets {
		sourceFile = filepath.Join(baseDir, s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = filepath.Join(baseDir, s.Destination)
		}

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		fmt.Println(" " + sourceFile + " -> " + destFile + " OK")
	}

	return nil
}
