/*
Copyright (C) 2020-2025  Daniele Rondina <geaaru@macaronios.org>
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

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
)

type CompilerOpts struct {
	Sources        []string
	GroupsEnabled  []string
	GroupsDisabled []string
}

func (o *CompilerOpts) IsGroupEnabled(g string) bool {
	ans := true

	if len(o.GroupsDisabled) == 0 && len(o.GroupsEnabled) == 0 {
		return ans
	}

	if len(o.GroupsEnabled) > 0 {
		ans := false

		for _, name := range o.GroupsEnabled {
			if name == g {
				ans = true
				break
			}
		}
		if !ans {
			return ans
		}
	}

	if len(o.GroupsDisabled) > 0 {
		for _, name := range o.GroupsDisabled {
			if name == g {
				ans = false
				break
			}
		}
	}

	return ans
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

		if !opts.IsGroupEnabled(group.Name) {
			continue
		}

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

	envBaseAbs, err := filepath.Abs(compiler.GetEnvBaseDir())
	if err != nil {
		return err
	}

	// Set node key with current group
	(*compiler.GetVars())["group"] = group

	for _, s := range targets {
		sourceFile = filepath.Join(envBaseAbs, s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = filepath.Join(envBaseAbs, s.Destination)
		}

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		log.GetDefaultLogger().Info(" " + sourceFile + " -> " + destFile + " OK")
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

	envBaseAbs, err := filepath.Abs(compiler.GetEnvBaseDir())
	if err != nil {
		return err
	}

	for _, s := range targets {
		sourceFile = filepath.Join(envBaseAbs, s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = filepath.Join(envBaseAbs, s.Destination)
		}

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		log.GetDefaultLogger().Info(" " + sourceFile + " -> " + destFile + " OK")
	}

	return nil
}

func CompileNodeFiles(node specs.LxdCNode, compiler LxdCTemplateCompiler, opts CompilerOpts) error {
	var sourceFile, destFile, baseDir string
	var targets []specs.LxdCConfigTemplate = []specs.LxdCConfigTemplate{}
	logger := log.GetDefaultLogger()

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

	logger.InfoC(logger.Aurora.Bold(
		logger.Aurora.BrightCyan(
			fmt.Sprintf(">>> [%s] Compile %d resources... :icecream:", node.GetName(), len(targets)))))

	// Set node key with current node
	(*compiler.GetVars())["node"] = node

	if len(node.Labels) > 0 {
		for k, v := range node.Labels {
			(*compiler.GetVars())[k] = v
		}
	}

	envBaseAbs, err := filepath.Abs(compiler.GetEnvBaseDir())
	if err != nil {
		return err
	}

	if filepath.IsAbs(node.SourceDir) {
		baseDir, err = filepath.Abs(node.SourceDir)
		if err != nil {
			return err
		}
	} else {
		baseDir = filepath.Join(envBaseAbs, node.SourceDir)
	}

	for idx, s := range targets {
		sourceFile = filepath.Join(baseDir, s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = filepath.Join(baseDir, s.Destination)
		}

		logger.DebugC(
			logger.Aurora.Italic(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] Compiling %s -> %s :coffee:",
						node.GetName(), sourceFile, destFile))))

		err := compiler.Compile(sourceFile, destFile)
		if err != nil {
			return err
		}

		logger.InfoC(
			logger.Aurora.BrightCyan(
				fmt.Sprintf(">>> [%s] - [%2d/%2d] %s :check_mark:",
					node.GetName(), idx+1, len(targets), destFile)))
	}

	return nil
}
