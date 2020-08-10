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
package loader

import (
	"errors"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"

	log "github.com/MottainaiCI/lxd-compose/pkg/logger"
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	helpers "github.com/mudler/luet/pkg/helpers"
)

type LxdCInstance struct {
	Config       *specs.LxdComposeConfig
	Logger       *log.LxdCLogger
	Environments []specs.LxdCEnvironment
}

func NewLxdCInstance(config *specs.LxdComposeConfig) *LxdCInstance {
	ans := &LxdCInstance{
		Config:       config,
		Logger:       log.NewLxdCLogger(config),
		Environments: make([]specs.LxdCEnvironment, 0),
	}

	// Initialize logging
	if config.GetLogging().EnableLogFile && config.GetLogging().Path != "" {
		err := ans.Logger.InitLogger2File()
		if err != nil {
			ans.Logger.Fatal("Error on initialize logfile")
		}
	}
	ans.Logger.SetAsDefault()

	return ans
}

func (i *LxdCInstance) AddEnvironment(env specs.LxdCEnvironment) {
	i.Environments = append(i.Environments, env)
}

func (i *LxdCInstance) GetEnvironments() *[]specs.LxdCEnvironment {
	return &i.Environments
}

func (i *LxdCInstance) GetEnvByProjectName(name string) *specs.LxdCEnvironment {
	for _, e := range i.Environments {
		for _, p := range e.Projects {
			if p.Name == name {
				return &e
			}
		}
	}

	return nil
}

func (i *LxdCInstance) GetEntitiesByNodeName(name string) (*specs.LxdCEnvironment, *specs.LxdCProject, *specs.LxdCGroup, *specs.LxdCNode) {
	for _, e := range i.Environments {
		for _, p := range e.Projects {
			for _, g := range p.Groups {
				for _, n := range g.Nodes {
					if n.Name == name {
						return &e, &p, &g, &n
					}
				}
			}
		}
	}
	return nil, nil, nil, nil
}

func (i *LxdCInstance) GetConfig() *specs.LxdComposeConfig {
	return i.Config
}

func (i *LxdCInstance) Validate(ignoreError bool) error {
	var ans error = nil
	mproj := make(map[string]int, 0)
	mnodes := make(map[string]int, 0)
	mgroups := make(map[string]int, 0)
	dupProjs := 0
	dupNodes := 0
	dupGroups := 0
	wrongHooks := 0

	// Check for duplicated project name
	for _, env := range i.Environments {

		for _, proj := range env.Projects {

			if _, isPresent := mproj[proj.Name]; isPresent {
				if !ignoreError {
					return errors.New("Duplicated project " + proj.Name)
				}

				i.Logger.Warning("Found duplicated project " + proj.Name)

				dupProjs++

			} else {
				mproj[proj.Name] = 1
			}

			// Check project's hooks events
			for _, h := range proj.Hooks {
				if (h.Event == "pre-project" || h.Event == "pre-group") && h.Node != "host" {
					i.Logger.Warning("On project " + proj.Name + " is present an hook " +
						h.Event + " for node " + h.Node + ". Only node host is admitted.")

					wrongHooks++

					if !ignoreError {
						return errors.New("Invalid hook for node " + h.Node +
							" on project " + proj.Name)
					}

				}

			}

			// Check groups
			for _, grp := range proj.Groups {

				if _, isPresent := mgroups[grp.Name]; isPresent {
					if !ignoreError {
						return errors.New("Duplicated group " + grp.Name)
					}

					i.Logger.Warning("Found duplicated group " + grp.Name)

					dupGroups++

				} else {
					mgroups[grp.Name] = 1
				}

				// Check group's hooks events
				if len(grp.Hooks) > 0 {
					for _, h := range grp.Hooks {
						if h.Event != "pre-node-creation" &&
							h.Event != "post-node-creation" &&
							h.Event != "pre-node-sync" &&
							h.Event != "post-node-sync" &&
							h.Event != "pre-group" &&
							h.Event != "post-group" {

							wrongHooks++

							i.Logger.Warning("Found invalid hook of type " + h.Event +
								" on group " + grp.Name)

							if !ignoreError {
								return errors.New("Invalid hook " + h.Event + " on group " + grp.Name)
							}
						}

					}
				}

				for _, node := range grp.Nodes {

					if _, isPresent := mnodes[node.Name]; isPresent {
						if !ignoreError {
							return errors.New("Duplicated node " + node.Name)
						}

						i.Logger.Warning("Found duplicated node " + node.Name)

						dupNodes++

					} else {
						mnodes[node.Name] = 1
					}

					if len(node.Hooks) > 0 {
						for _, h := range node.Hooks {
							if h.Node != "" {
								i.Logger.Warning("Invalid hook on node " + node.Name + " with node field valorized.")
								wrongHooks++
								if !ignoreError {
									return errors.New("Invalid hook on node " + node.Name)
								}
							}

							if h.Event != "pre-node-creation" &&
								h.Event != "post-node-creation" &&
								h.Event != "pre-node-sync" &&
								h.Event != "post-node-sync" {

								wrongHooks++

								i.Logger.Warning("Found invalid hook of type " + h.Event +
									" on node " + node.Name)

								if !ignoreError {
									return errors.New("Invalid hook " + h.Event + " on node " + node.Name)
								}
							}
						}

					}

				}

			}
		}

		return nil
	}

	return ans
}

func (i *LxdCInstance) LoadEnvironments() error {
	var regexConfs = regexp.MustCompile(`.yml$`)

	if len(i.Config.GetEnvironmentDirs()) == 0 {
		return errors.New("No environment directories configured.")
	}

	for _, edir := range i.Config.GetEnvironmentDirs() {
		i.Logger.Debug("Checking directory", edir, "...")

		files, err := ioutil.ReadDir(edir)
		if err != nil {
			i.Logger.Debug("Skip dir", edir, ":", err.Error())
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if !regexConfs.MatchString(file.Name()) {
				i.Logger.Debug("File", file.Name(), "skipped.")
				continue
			}

			content, err := ioutil.ReadFile(path.Join(edir, file.Name()))
			if err != nil {
				i.Logger.Debug("On read file", file.Name(), ":", err.Error())
				i.Logger.Debug("File", file.Name(), "skipped.")
				continue
			}

			env, err := specs.EnvironmentFromYaml(content, path.Join(edir, file.Name()))
			if err != nil {
				i.Logger.Debug("On parse file", file.Name(), ":", err.Error())
				i.Logger.Debug("File", file.Name(), "skipped.")
				continue
			}

			err = i.loadExtraFiles(env)

			i.AddEnvironment(*env)

			i.Logger.Debug("Loaded environment file " + env.File)

		}

	}

	return nil
}

func (i *LxdCInstance) loadExtraFiles(env *specs.LxdCEnvironment) error {
	envBaseDir, err := filepath.Abs(path.Dir(env.File))
	if err != nil {
		return err
	}

	for idx, proj := range env.Projects {

		// Load external groups files
		if len(proj.IncludeGroupFiles) > 0 {

			// Load external groups files
			for _, gfile := range proj.IncludeGroupFiles {

				if !helpers.Exists(path.Join(envBaseDir, gfile)) {
					i.Logger.Warning("For project", proj.Name, "included group file", gfile,
						"is not present.")
					continue
				}

				content, err := ioutil.ReadFile(path.Join(envBaseDir, gfile))
				if err != nil {
					i.Logger.Debug("On read file", gfile, ":", err.Error())
					i.Logger.Debug("File", gfile, "skipped.")
					continue
				}

				grp, err := specs.GroupFromYaml(content)
				if err != nil {
					i.Logger.Debug("On parse file", gfile, ":", err.Error())
					i.Logger.Debug("File", gfile, "skipped.")
					continue
				}

				env.Projects[idx].AddGroup(grp)
			}

		}

		if len(proj.IncludeEnvFiles) > 0 {
			// Load external env vars files
			for _, efile := range proj.IncludeEnvFiles {

				if !helpers.Exists(path.Join(envBaseDir, efile)) {
					i.Logger.Warning("For project", proj.Name, "included env file", efile,
						"is not present.")
					continue
				}

				i.Logger.Debug("Loaded environment file " + env.File)

				if path.Ext(efile) != ".yml" {
					i.Logger.Warning("For project", proj.Name, "included env file", efile,
						"will be used only with template compiler")
					continue
				}

				content, err := ioutil.ReadFile(path.Join(envBaseDir, efile))
				if err != nil {
					i.Logger.Debug("On read file", efile, ":", err.Error())
					i.Logger.Debug("File", efile, "skipped.")
					continue
				}

				evars, err := specs.EnvVarsFromYaml(content)
				if err != nil {
					i.Logger.Debug("On parse file", efile, ":", err.Error())
					i.Logger.Debug("File", efile, "skipped.")
					continue
				}

				env.Projects[idx].AddEnvironment(evars)

			}

		}

	}

	return nil
}
