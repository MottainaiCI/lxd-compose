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
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"

	helpers "github.com/mudler/luet/pkg/helpers"
)

type LxdCInstance struct {
	Config       *specs.LxdComposeConfig
	Environments []specs.LxdCEnvironment
}

func NewLxdCInstance(config *specs.LxdComposeConfig) *LxdCInstance {
	return &LxdCInstance{
		Config:       config,
		Environments: make([]specs.LxdCEnvironment, 0),
	}
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

	// Check for duplicated project name
	for _, env := range i.Environments {

		for _, proj := range env.Projects {

			if _, isPresent := mproj[proj.Name]; isPresent {
				if !ignoreError {
					return errors.New("Duplicated project " + proj.Name)
				}

				fmt.Println("Found duplicated project " + proj.Name)

				dupProjs++

			} else {
				mproj[proj.Name] = 1
			}

			// Check groups

			for _, grp := range proj.Groups {

				if _, isPresent := mgroups[grp.Name]; isPresent {
					if !ignoreError {
						return errors.New("Duplicated group " + grp.Name)
					}

					fmt.Println("Found duplicated group " + grp.Name)

					dupGroups++

				} else {
					mgroups[grp.Name] = 1
				}

				for _, node := range grp.Nodes {

					if _, isPresent := mnodes[node.Name]; isPresent {
						if !ignoreError {
							return errors.New("Duplicated node " + node.Name)
						}

						fmt.Println("Found duplicated node " + node.Name)

						dupNodes++

					} else {
						mnodes[node.Name] = 1
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
		fmt.Println("Checking directory", edir, "...")

		files, err := ioutil.ReadDir(edir)
		if err != nil {
			fmt.Println("Skip dir", edir, ":", err.Error())
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if !regexConfs.MatchString(file.Name()) {
				fmt.Println("File", file.Name(), "skipped.")
				continue
			}

			content, err := ioutil.ReadFile(path.Join(edir, file.Name()))
			if err != nil {
				fmt.Println("On read file", file.Name(), ":", err.Error())
				fmt.Println("File", file.Name(), "skipped.")
				continue
			}

			env, err := specs.EnvironmentFromYaml(content, path.Join(edir, file.Name()))
			if err != nil {
				fmt.Println("On parse file", file.Name(), ":", err.Error())
				fmt.Println("File", file.Name(), "skipped.")
				continue
			}

			err = loadExtraFiles(env)

			i.AddEnvironment(*env)

		}

	}

	return nil
}

func loadExtraFiles(env *specs.LxdCEnvironment) error {
	envBaseDir := path.Dir(env.File)

	for _, proj := range env.Projects {

		// Load external groups files
		if len(proj.IncludeGroupFiles) > 0 {

			// Load external groups files
			for _, gfile := range proj.IncludeGroupFiles {

				if !helpers.Exists(path.Join(envBaseDir, gfile)) {
					fmt.Println("For project", proj.Name, "included group file", gfile,
						"is not present.")
					continue
				}

				content, err := ioutil.ReadFile(path.Join(envBaseDir, gfile))
				if err != nil {
					fmt.Println("On read file", gfile, ":", err.Error())
					fmt.Println("File", gfile, "skipped.")
					continue
				}

				grp, err := specs.GroupFromYaml(content)
				if err != nil {
					fmt.Println("On parse file", gfile, ":", err.Error())
					fmt.Println("File", gfile, "skipped.")
					continue
				}

				proj.AddGroup(grp)
			}

			// Load external env vars files
			for _, efile := range proj.IncludeEnvFiles {

				if !helpers.Exists(path.Join(envBaseDir, efile)) {
					fmt.Println("For project", proj.Name, "included env file", efile,
						"is not present.")
					continue
				}

				if path.Ext(efile) != ".yml" {
					fmt.Println("For project", proj.Name, "included env file", efile,
						"will be used only with template compiler")
					continue
				}

				content, err := ioutil.ReadFile(path.Join(envBaseDir, efile))
				if err != nil {
					fmt.Println("On read file", efile, ":", err.Error())
					fmt.Println("File", efile, "skipped.")
					continue
				}

				evars, err := specs.EnvVarsFromYaml(content)
				if err != nil {
					fmt.Println("On parse file", efile, ":", err.Error())
					fmt.Println("File", efile, "skipped.")
					continue
				}

				proj.AddEnvironment(evars)
			}

		}

	}

	return nil
}
