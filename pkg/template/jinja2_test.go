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
package template_test

import (
	specs "github.com/MottainaiCI/lxd-compose/pkg/specs"
	. "github.com/MottainaiCI/lxd-compose/pkg/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("", func() {

	Context("TemplateJinja1", func() {

		proj := &specs.LxdCProject{
			Name: "project1",
			Environments: []specs.LxdCEnvVars{
				{
					EnvVars: map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
						"key3": map[string]string{
							"f1": "foo",
							"f2": "foo2",
						},
					},
				},
			},
		}

		c := NewJinja2Compiler(proj)
		c.InitVars()

		It("Compilation1", func() {

			sourceData := `
k1: "{{ key1 }}"
k2: "{{ key2 }}"
`
			out, err := c.CompileRaw(sourceData)

			expectedOutput := `
k1: "value1"
k2: "value2"
`
			Expect(err).Should(BeNil())
			Expect(out).To(Equal(expectedOutput))
		})

	})
})
