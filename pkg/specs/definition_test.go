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
package specs_test

import (
	. "github.com/MottainaiCI/lxd-compose/pkg/specs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Specs Test", func() {

	Context("Environment1", func() {

		e1 := []byte(`
version: "1"

template_engine:
  engine: "jinja2"
`)

		env1, err := EnvironmentFromYaml(e1, "file1.yml")

		It("Convert env1", func() {

			expected_env1 := &LxdCEnvironment{
				Version: "1",
				File:    "file1.yml",
				TemplateEngine: LxdCTemplateEngine{
					Engine: "jinja2",
				},
				Commands:             []LxdCCommand{},
				IncludeCommandsFiles: []string{},
			}

			Expect(err).Should(BeNil())
			Expect(env1).To(Equal(expected_env1))
		})

	})

	Context("Envs", func() {

		It("Convert env1", func() {
			env1 := []byte(`
envs:
  key1: "value1"
  key2: "value2"
`)

			e1, err := EnvVarsFromYaml(env1)
			expected_env1 := &LxdCEnvVars{
				EnvVars: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			}

			Expect(err).Should(BeNil())
			Expect(e1).To(Equal(expected_env1))
		})

	})

	Context("Group1", func() {

		g1 := []byte(`
name: "group1"
description: "description1"
connection: "local"
common_profiles:
- profile1
- profile2

ephemeral: false

hooks:
  - event: pre-node-creation
    commands:
     - echo 1
     - echo 2

nodes:
- name: "node1"
  image_source: "sabayon"
`)

		grp, err := GroupFromYaml(g1)

		It("Convert group1", func() {

			expected_grp := &LxdCGroup{
				Name:           "group1",
				Description:    "description1",
				Connection:     "local",
				CommonProfiles: []string{"profile1", "profile2"},
				Ephemeral:      false,
				Nodes: []LxdCNode{
					{
						Name:        "node1",
						ImageSource: "sabayon",
					},
				},
				Hooks: []LxdCHook{
					{
						Event: "pre-node-creation",
						Commands: []string{
							"echo 1",
							"echo 2",
						},
					},
				},
			}

			Expect(err).Should(BeNil())
			Expect(grp).To(Equal(expected_grp))
			Expect(len(grp.Hooks)).To(Equal(1))
		})

	})

	Context("Envs", func() {

		It("Convert env1", func() {
			env1 := []byte(`
envs:
  key1: "value1"
  key2: "value2"
`)

			e1, err := EnvVarsFromYaml(env1)
			expected_env1 := &LxdCEnvVars{
				EnvVars: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			}

			Expect(err).Should(BeNil())
			Expect(e1).To(Equal(expected_env1))
		})

		It("Convert env2", func() {
			env2 := []byte(`
envs:
  key1: "value1"
  key2:
   complex:
     arg2: 10
`)

			e2, err := EnvVarsFromYaml(env2)
			obj1 := e2.EnvVars["key2"]

			Expect(err).Should(BeNil())
			Expect(e2.EnvVars["key1"]).To(Equal("value1"))

			for k, v := range obj1.(map[string]interface{}) {
				Expect(k).To(Equal("complex"))
				for k2, v2 := range v.(map[string]interface{}) {
					Expect(k2).To(Equal("arg2"))
					Expect(v2).To(Equal(10))
				}

			}
		})
	})
})
