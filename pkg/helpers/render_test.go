package helpers

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Render Helm test unit", func() {

	Context("Render1", func() {

		cf1 := &chart.File{
			Name: "file1",
			Data: []byte(`
val: "1"
`),
		}

		cf2 := &chart.File{
			Name: "templates",
			Data: []byte(`
{{ .Values.v }}
{{- include "file1" . }}
`),
		}

		c := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "",
				Version: "",
			},
			Templates: []*chart.File{
				cf1, cf2,
			},
			Values: map[string]interface{}{
				"Values": map[string]string{
					"v": "valuev",
				},
			},
		}

		v, err := chartutil.CoalesceValues(c, map[string]interface{}{})

		out, err2 := engine.Render(c, v)

		It("render check1", func() {
			expectfile1 := `
val: "1"
`

			expectTemplates := `
valuev
val: "1"

`

			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())

			/*
				for k, v := range out {
					fmt.Println("KEY ", k)
					fmt.Println("V ", v)
				}
			*/

			Expect(out["templates"]).To(Equal(expectTemplates))
			Expect(out["file1"]).To(Equal(expectfile1))
		})
	})

	Context("Render2", func() {

		cf1 := &chart.File{
			Name: "file1",
			Data: []byte(`
val: "1"
`),
		}

		cf2 := &chart.File{
			Name: "file2",
			Data: []byte(`
{{- define "def1" }}
- test test test
{{- end }}
`),
		}

		cft := &chart.File{
			Name: "templates",
			Data: []byte(`
{{ .Values.v }}
{{- include "tpl/file1" . }}
{{- template "def1" . }}
`),
		}

		c := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "tpl",
				Version: "",
			},
			Templates: []*chart.File{
				cf1, cf2, cft,
			},
			Values: map[string]interface{}{
				"Values": map[string]string{
					"v": "valuev",
				},
			},
		}

		v, err := chartutil.CoalesceValues(c, map[string]interface{}{
			"Values": map[string]string{
				"v": "value2",
			},
		})

		out, err2 := engine.Render(c, v)

		expectfile1 := `
val: "1"
`

		expectfile2 := `
`
		expectTemplates := `
value2
val: "1"

- test test test
`
		It("render check2", func() {

			Expect(err).Should(BeNil())
			Expect(err2).Should(BeNil())

			/*
				for k, v := range out {
					fmt.Println("KEY ", k)
					fmt.Println("V ", v)
				}
			*/

			Expect(out["tpl/templates"]).To(Equal(expectTemplates))
			Expect(out["tpl/file1"]).To(Equal(expectfile1))
			Expect(out["tpl/file2"]).To(Equal(expectfile2))
		})
	})

})
