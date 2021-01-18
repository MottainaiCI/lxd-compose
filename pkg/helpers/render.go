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
package helpers

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

func RenderContent(raw, valuesFile, defaultFile, originFile string) (string, error) {
	if !Exists(valuesFile) {
		return "", errors.New(fmt.Sprintf(
			"Render value file %s not existing ", valuesFile))
	}
	val, err := ioutil.ReadFile(valuesFile)
	if err != nil {
		return "", errors.New(fmt.Sprintf(
			"Error on reading Render value file %s: %s", valuesFile, err.Error()))
	}

	var values map[string]interface{}
	d := make(map[string]interface{}, 0)
	if defaultFile != "" {
		if !Exists(defaultFile) {
			return "", errors.New(fmt.Sprintf(
				"Render value file %s not existing ", defaultFile))
		}

		def, err := ioutil.ReadFile(defaultFile)
		if err != nil {
			return "", errors.New(fmt.Sprintf(
				"Error on reading Render value file %s: %s", valuesFile, err.Error()))
		}

		if err = yaml.Unmarshal(def, &d); err != nil {
			return "", errors.New(fmt.Sprintf(
				"Error on unmarshal file %s: %s", defaultFile, err.Error()))
		}
	}

	if err = yaml.Unmarshal(val, &values); err != nil {
		return "", errors.New(fmt.Sprintf(
			"Error on unmarsh file %s: %s", valuesFile, err.Error()))
	}

	c := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "",
			Version: "",
		},
		Templates: []*chart.File{
			{Name: "templates", Data: []byte(raw)},
		},
		Values: map[string]interface{}{"Values": values},
	}

	v, err := chartutil.CoalesceValues(c, map[string]interface{}{"Values": d})
	if err != nil {
		return "", errors.New(fmt.Sprintf(
			"Error on coalesce values for file %s: %s", originFile, err.Error()))
	}
	out, err := engine.Render(c, v)
	if err != nil {
		return "", errors.New(fmt.Sprintf(
			"Error on rendering file %s: %s", originFile, err.Error()))
	}

	return out["templates"], nil
}
