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
package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	tarf "github.com/geaaru/tar-formers/pkg/executor"
	tarf_specs "github.com/geaaru/tar-formers/pkg/specs"
	tools "github.com/geaaru/tar-formers/pkg/tools"
)

type PackJob struct {
	Envs                []*specs.LxdCEnvironment
	MapHooks            map[string]bool
	MapTemplates        map[string]bool
	MapEnvsfiles        map[string]bool
	MapExtra            map[string]bool
	MapEIncludes        map[string]bool
	MapRenames          map[string]bool
	Specfile            *tarf_specs.SpecFile
	SourceDir           string
	TargetSourcedir     string
	SourceCommonPathDir string

	EnvsDirs map[string]bool
}

func (j *PackJob) HasEnv(e *specs.LxdCEnvironment) bool {
	for idx := range j.Envs {
		if j.Envs[idx] == e {
			return true
		}
	}
	return false
}

func (i *LxdCInstance) preparePackSpec4Group(
	envBaseDir string, job *PackJob, g *specs.LxdCGroup,
	templates *[]specs.LxdCConfigTemplate) {

	// Load hooks from group
	if len(g.IncludeHooksFiles) > 0 {
		for _, f := range g.IncludeHooksFiles {

			hfile := filepath.Join(envBaseDir, f)
			if _, present := job.MapHooks[hfile]; !present {
				i.Logger.Debug(fmt.Sprintf(
					"Inject of the group hook %s", f))
				job.MapHooks[hfile] = true
				job.Specfile.Writer.AddFile(hfile)
			}
		}
	}

	// Check the nodes
	for _, n := range g.Nodes {
		// Load hooks from node
		if len(n.IncludeHooksFiles) > 0 {
			for _, f := range n.IncludeHooksFiles {
				hfile := filepath.Join(envBaseDir, f)
				if _, present := job.MapHooks[hfile]; !present {
					i.Logger.Debug(fmt.Sprintf(
						"Inject of the node hook %s", f))
					job.MapHooks[hfile] = true
					job.Specfile.Writer.AddFile(hfile)
				}
			}
		}

		// Load templates files from groups and project
		if len(*templates) > 0 {
			for _, t := range *templates {
				tfile := filepath.Join(envBaseDir, t.Source)
				if _, present := job.MapTemplates[tfile]; !present {
					job.MapTemplates[tfile] = true
					job.Specfile.Writer.AddFile(tfile)

					if job.SourceCommonPathDir != "" && strings.HasPrefix(t.Source, job.SourceCommonPathDir) {
						ntfile := filepath.Join(job.SourceDir,
							t.Source[len(job.SourceCommonPathDir):])
						i.Logger.InfoC(fmt.Sprintf("Template %s -> %s",
							tfile, ntfile))
						job.Specfile.Rename = append(job.Specfile.Rename,
							tarf_specs.RenameRule{
								Source: tfile,
								Dest:   ntfile,
							},
						)
					}
				}
			}
		}

		// Load templates files from node
		if len(n.ConfigTemplates) > 0 {
			for _, t := range n.ConfigTemplates {
				tfile := filepath.Join(envBaseDir, t.Source)
				if _, present := job.MapTemplates[tfile]; !present {
					job.MapTemplates[tfile] = true
					job.Specfile.Writer.AddFile(tfile)

					if job.SourceCommonPathDir != "" && strings.HasPrefix(t.Source, job.SourceCommonPathDir) {
						ntfile := filepath.Join(job.SourceDir+"/",
							t.Source[len(job.SourceCommonPathDir):])
						i.Logger.InfoC(fmt.Sprintf("Template %s -> %s",
							tfile, ntfile))
						job.Specfile.Rename = append(job.Specfile.Rename,
							tarf_specs.RenameRule{
								Source: tfile,
								Dest:   ntfile,
							},
						)
					}
				}
			}
		}

	}
}

func preparePackSpec4Command(
	envBaseDir string, job *PackJob, c *specs.LxdCCommand) {
	// Load var files
	if len(c.VarFiles) > 0 {
		for _, f := range c.VarFiles {
			vfile := filepath.Join(envBaseDir, f)
			if _, present := job.MapEnvsfiles[vfile]; !present {
				job.MapEnvsfiles[vfile] = true
				job.Specfile.Writer.AddFile(vfile)
			}
		}
	}
	// Load hooks
	if len(c.IncludeHooksFiles) > 0 {
		for _, f := range c.IncludeHooksFiles {
			pfile := filepath.Join(envBaseDir, f)
			// Hooks could be common between different projects.
			if _, present := job.MapHooks[pfile]; !present {
				job.MapHooks[pfile] = true
				job.Specfile.Writer.AddFile(pfile)
			}
		}
	}
}

func (i *LxdCInstance) preparePackSpec4Project(job *PackJob, project string) error {
	e := i.GetEnvByProjectName(project)
	if e == nil {
		return fmt.Errorf("No environment found for project %s.", project)
	}

	job.Specfile.Writer.AddFile(e.File)
	i.Logger.InfoC(fmt.Sprintf(
		":factory:Processing project %s with env file %s.", project, e.File))

	envBaseDir := filepath.Dir(e.File)

	job.EnvsDirs[envBaseDir] = true

	// Calculate the target source dir only the first time
	if job.TargetSourcedir == "" && job.SourceDir != "" {
		job.TargetSourcedir = job.SourceDir
		n := len(strings.Split(envBaseDir, "/"))
		for i := 0; i < n; i++ {
			job.TargetSourcedir = "../" + job.TargetSourcedir
		}

	}

	// Check if the environment is already been added
	if !job.HasEnv(e) {

		// NOTE: For includes the path is always based on envBaseDir.

		// Check include of the hooks and vars files from commands
		if len(e.Commands) > 0 {
			for idx := range e.Commands {
				preparePackSpec4Command(envBaseDir, job, &e.Commands[idx])
			}
		}

		// Add commands files
		if len(e.IncludeCommandsFiles) > 0 {
			for _, f := range e.IncludeCommandsFiles {
				cfile := filepath.Join(envBaseDir, f)
				if _, present := job.MapEIncludes[cfile]; !present {
					job.MapEIncludes[cfile] = true
					job.Specfile.Writer.AddFile(cfile)

					// Load command file to read hooks

					content, err := os.ReadFile(cfile)
					if err != nil {
						return fmt.Errorf(
							"Error on read command file %s: %s", cfile, err.Error())
					}

					if i.Config.IsEnableRenderEngine() {
						// Render file
						renderOut, err := helpers.RenderContent(string(content),
							i.Config.RenderValuesFile,
							i.Config.RenderDefaultFile,
							cfile,
							i.Config.RenderEnvsVars,
						)
						if err != nil {
							return err
						}

						content = []byte(renderOut)
					}

					cmd, err := specs.CommandFromYaml(content)
					if err != nil {
						return fmt.Errorf(
							"Error on parse command file %s: %s", cfile, err.Error())
					}

					preparePackSpec4Command(envBaseDir, job, cmd)
					cmd = nil
				}
			}
		}

		// Add LXD profiles files
		if len(e.IncludeProfilesFiles) > 0 {
			for _, f := range e.IncludeProfilesFiles {
				fpath := filepath.Join(envBaseDir, f)
				if _, present := job.MapEIncludes[fpath]; !present {
					job.MapEIncludes[fpath] = true
					job.Specfile.Writer.AddFile(fpath)
				}
			}
		}

		// Add LXD networks files
		if len(e.IncludeNetworkFiles) > 0 {
			for _, f := range e.IncludeNetworkFiles {
				fpath := filepath.Join(envBaseDir, f)
				if _, present := job.MapEIncludes[fpath]; !present {
					job.MapEIncludes[fpath] = true
					job.Specfile.Writer.AddFile(fpath)
				}
			}
		}

		// Add LXD storages files
		if len(e.IncludeStorageFiles) > 0 {
			for _, f := range e.IncludeStorageFiles {
				fpath := filepath.Join(envBaseDir, f)
				if _, present := job.MapEIncludes[fpath]; !present {
					job.MapEIncludes[fpath] = true
					job.Specfile.Writer.AddFile(fpath)
				}
			}
		}

		// Add LXD ACL files
		if len(e.IncludeAclsFiles) > 0 {
			for _, f := range e.IncludeAclsFiles {
				fpath := filepath.Join(envBaseDir, f)
				if _, present := job.MapEIncludes[fpath]; !present {
					job.MapEIncludes[fpath] = true
					job.Specfile.Writer.AddFile(fpath)
				}
			}
		}

		// Check for extra pack files
		if e.PackExtra != nil {
			if len(e.PackExtra.Dirs) > 0 {
				for _, d := range e.PackExtra.Dirs {
					dpath := filepath.Join(envBaseDir, d)
					if _, present := job.MapEIncludes[dpath]; !present {
						job.MapEIncludes[dpath] = true
						job.Specfile.Writer.AddDir(dpath)
					}
				}
			}
			if len(e.PackExtra.Files) > 0 {
				for _, f := range e.PackExtra.Files {
					fpath := filepath.Join(envBaseDir, f)
					if _, present := job.MapEIncludes[fpath]; !present {
						job.MapEIncludes[fpath] = true
						job.Specfile.Writer.AddFile(fpath)
					}
				}
			}

			if len(e.PackExtra.Rename) > 0 {
				for _, r := range e.PackExtra.Rename {
					fsource := r.Source
					if _, present := job.MapRenames[fsource]; !present {
						job.MapRenames[fsource] = true
						job.Specfile.Rename = append(
							job.Specfile.Rename, *r,
						)
					}
				}

			}
		}

	}

	p := e.GetProjectByName(project)

	// The templates could be based on source dir of the node.
	// I postpone the elaboration when I'm at node level.
	templates := []specs.LxdCConfigTemplate{}
	if len(p.ConfigTemplates) > 0 {
		templates = append(templates, p.ConfigTemplates...)
	}

	// Load project hooks
	if len(p.IncludeHooksFiles) > 0 {
		for _, f := range p.IncludeHooksFiles {
			pfile := filepath.Join(envBaseDir, f)
			// Hooks could be common between different projects.
			if _, present := job.MapHooks[pfile]; !present {
				job.MapHooks[pfile] = true
				job.Specfile.Writer.AddFile(pfile)
			}
		}
	}

	// Load var files
	if len(p.IncludeEnvFiles) > 0 {
		for _, f := range p.IncludeEnvFiles {
			vfile := filepath.Join(envBaseDir, f)
			// Variables files could be common between different projects.
			if _, present := job.MapEnvsfiles[vfile]; !present {
				job.MapEnvsfiles[vfile] = true
				job.Specfile.Writer.AddFile(vfile)
			}
		}
	}

	// Using map to exclude groups processed by includes
	grpMap := make(map[string]bool, 0)

	// Load project files
	if len(p.IncludeGroupFiles) > 0 {
		for _, f := range p.IncludeGroupFiles {
			gfile := filepath.Join(envBaseDir, f)

			job.Specfile.Writer.AddFile(gfile)

			// NOTE: I can't retrieve the group from the file.
			//       So I need read again the file directly.
			//       PRE: For pack operation the specs must be clean
			//       without broken includes.

			content, err := os.ReadFile(gfile)
			if err != nil {
				return fmt.Errorf(
					"Error on read group file %s: %s", gfile, err.Error())
			}

			if i.Config.IsEnableRenderEngine() {
				// Render file
				renderOut, err := helpers.RenderContent(string(content),
					i.Config.RenderValuesFile,
					i.Config.RenderDefaultFile,
					gfile,
					i.Config.RenderEnvsVars,
				)
				if err != nil {
					return err
				}

				content = []byte(renderOut)
			}

			grp, err := specs.GroupFromYaml(content)
			if err != nil {
				return fmt.Errorf(
					"Error on parse group file %s: %s", gfile, err.Error())
			}

			// Get group loaded
			g := p.GetGroupByName(grp.Name)
			if g == nil {
				return fmt.Errorf(
					"Unexpected state on retrieve group %s.", grp.Name)
			}
			grpMap[grp.Name] = true

			tgroup := append(templates, grp.ConfigTemplates...)
			grp = nil

			i.preparePackSpec4Group(envBaseDir, job, g, &tgroup)
		}
	}

	// Processing all groups not included by file
	for _, g := range *p.GetGroups() {
		if _, present := grpMap[g.Name]; !present {
			tgroup := append(templates, g.ConfigTemplates...)

			i.preparePackSpec4Group(envBaseDir, job, &g, &tgroup)
		}
	}

	return nil
}

func (i *LxdCInstance) PackProjects(tarball, sourceCPDir string,
	projects []string) (string, error) {

	// Prepare tar-formers specs
	s := tarf_specs.NewSpecFile()
	s.SameChtimes = true
	s.Writer = tarf_specs.NewWriter()

	job := &PackJob{
		Envs:                []*specs.LxdCEnvironment{},
		MapHooks:            make(map[string]bool, 0),
		MapEnvsfiles:        make(map[string]bool, 0),
		MapTemplates:        make(map[string]bool, 0),
		MapEIncludes:        make(map[string]bool, 0),
		MapExtra:            make(map[string]bool, 0),
		EnvsDirs:            make(map[string]bool, 0),
		MapRenames:          make(map[string]bool, 0),
		Specfile:            s,
		TargetSourcedir:     "",
		SourceCommonPathDir: sourceCPDir,
	}

	if sourceCPDir != "" {
		job.SourceDir = "sources"
	}

	// Using a map to avoid process multiple times the same
	// project if the user does errors.
	mapp := make(map[string]bool, 0)

	// Add directories and files to the tar-formers spec file
	for _, p := range projects {
		if _, ok := mapp[p]; !ok {
			mapp[p] = true

			err := i.preparePackSpec4Project(job, p)
			if err != nil {
				return "", err
			}
		}
	}

	// Add lxd-conf directory if defined.
	if i.Config.GetGeneral().LxdConfDir != "" {
		job.Specfile.Writer.AddDir(i.Config.GetGeneral().LxdConfDir)
	}
	if i.Config.RenderDefaultFile != "" {
		job.Specfile.Writer.AddFile(i.Config.RenderDefaultFile)
	}
	if i.Config.RenderValuesFile != "" {
		job.Specfile.Writer.AddFile(i.Config.RenderValuesFile)
	}

	if len(i.Config.RenderTemplatesDirs) > 0 {
		for _, d := range i.Config.RenderTemplatesDirs {
			job.Specfile.Writer.AddDir(d)
		}
	}

	// Create the new lxd-compose config file with only
	// the environment directories of the projects injected.
	cfg := i.Config.Clone()
	// Always disable debug by default
	cfg.General.Debug = false
	// Set only the selected env dir
	cfg.EnvironmentDirs = []string{}
	for k := range job.EnvsDirs {
		cfg.EnvironmentDirs = append(cfg.EnvironmentDirs, k)
	}
	// Create temporary file
	newCfgFile, err := os.CreateTemp("", "lxd-compose-pack*")
	if err != nil {
		return "", err
	}
	defer os.Remove(newCfgFile.Name())

	// Write the config
	data, err := cfg.Yaml()
	if err != nil {
		newCfgFile.Close()
		return "", err
	}
	if _, err := newCfgFile.Write(data); err != nil {
		newCfgFile.Close()
		return "", err
	}
	if err := newCfgFile.Close(); err != nil {
		return "", err
	}
	// Add the file for injection and his rename
	job.Specfile.Writer.AddFile(newCfgFile.Name())
	job.Specfile.Rename = append(job.Specfile.Rename,
		tarf_specs.RenameRule{
			Source: newCfgFile.Name(),
			Dest:   ".lxd-compose.yml",
		},
	)

	// Reduce memory usage
	job.MapHooks = nil
	job.MapEnvsfiles = nil
	job.MapEnvsfiles = nil
	job.MapEIncludes = nil
	job.MapTemplates = nil
	job.MapRenames = nil

	// Prepare tarball creation
	opts := tools.NewTarCompressionOpts(true)
	defer opts.Close()
	err = tools.PrepareTarWriter(tarball, opts)
	if err != nil {
		return "", err
	}

	tarformers := tarf.NewTarFormers(tarf.GetOptimusPrime().Config)
	if opts.CompressWriter != nil {
		tarformers.SetWriter(opts.CompressWriter)
	} else {
		tarformers.SetWriter(opts.FileWriter)
	}

	err = tarformers.RunTaskWriter(job.Specfile)
	if err != nil {
		return "", err
	}

	// The sources directory is set to the lxd-compose path
	// and not to envs.
	return job.TargetSourcedir, nil
}
