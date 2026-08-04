/*
Copyright © 2020-2026 Daniele Rondina <geaaru@macaronios.org>
See AUTHORS and LICENSE for the license details and contributors.
*/
package template

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MottainaiCI/lxd-compose/pkg/helpers"
	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	"golang.org/x/sync/semaphore"
)

type CompilerOpts struct {
	Sources        []string
	GroupsEnabled  []string
	GroupsDisabled []string
	Concurrency    int
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

	waitGroup := &sync.WaitGroup{}
	sem := semaphore.NewWeighted(int64(opts.Concurrency))
	ctx := context.TODO()
	var ch chan helpers.ChannelError = make(
		chan helpers.ChannelError,
		opts.Concurrency,
	)

	for _, s := range targets {
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

		waitGroup.Add(1)
		go compileRouting(compiler, ch, sem, waitGroup,
			sourceFile, destFile, &ctx)
	}

	nTargets := len(targets)
	fail := false
	for i, s := range targets {
		sourceFile = filepath.Join(baseDir, s.Source)
		if filepath.IsAbs(s.Destination) {
			destFile = s.Destination
		} else {
			destFile = filepath.Join(baseDir, s.Destination)
		}
		logger.DebugC(
			logger.Aurora.Italic(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] Waiting %s -> %s :coffee:",
						node.GetName(), sourceFile, destFile))))
		resp := <-ch
		if resp.Error != nil {
			logger.Error(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] - [%2d/%2d] %s - %s :cross_mark:",
						node.GetName(), i+1, nTargets, destFile, resp.Error.Error())))
			fail = true
		} else {
			logger.InfoC(
				logger.Aurora.BrightCyan(
					fmt.Sprintf(">>> [%s] - [%2d/%2d] %s :check_mark:",
						node.GetName(), i+1, nTargets, destFile)))
		}
	}

	waitGroup.Wait()

	if fail {
		return fmt.Errorf("errors on compile template files")
	}

	return nil
}

func compileRouting(compiler LxdCTemplateCompiler,
	channel chan helpers.ChannelError,
	sem *semaphore.Weighted, waitGroup *sync.WaitGroup,
	sourceFile, destFile string, ctx *context.Context) {

	defer waitGroup.Done()
	err := sem.Acquire(*ctx, 1)
	if err != nil {
		channel <- helpers.ChannelError{
			Error:   fmt.Errorf("error on acquire semaphore: %s", err.Error()),
			Closure: destFile,
		}
		return
	}
	defer sem.Release(1)

	err = compiler.Compile(sourceFile, destFile)
	if err != nil {
		channel <- helpers.ChannelError{
			Error:   err,
			Closure: destFile,
		}
		return
	}

	channel <- helpers.ChannelError{
		Error:   nil,
		Closure: destFile,
	}
	return
}
