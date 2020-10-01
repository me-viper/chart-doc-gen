/*
Copyright The Kubepack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	"kubepack.dev/chart-doc-gen/templates"

	"github.com/olekukonko/tablewriter"
	flag "github.com/spf13/pflag"
	y2 "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	docFile    *string = flag.StringP("doc", "d", "doc.yaml", "Path to a project's doc.{json|yaml} info file")
	valuesFile *string = flag.StringP("values", "v", "values.yaml", "Path to chart values file")
	tplFile    *string = flag.StringP("template", "t", "readme2.tpl", "Path to a doc template file")
	outFile    *string = flag.StringP("output", "o", "", "Path to a output file")
)

func main() {
	flag.Parse()

	f, err := os.Open(*docFile)
	if err != nil {
		panic(err)
	}
	reader := y2.NewYAMLOrJSONDecoder(f, 2048)
	var doc DocInfo
	err = reader.Decode(&doc)
	if err != nil && err != io.EOF {
		panic(err)
	}

	data, err := ioutil.ReadFile(*valuesFile)
	if err != nil {
		panic(err)
	}
	obj, err := yaml.Parse(string(data))
	if err != nil {
		panic(err)
	}
	rows, err := GenerateValuesTable(obj)
	if err != nil {
		panic(err)
	}

	var params [][]string
	for _, row := range rows {
		params = append(params, []string{
			row[0],
			row[1],
			fmt.Sprintf("`%s`", row[2]),
		})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Parameter", "Description", "Default"})
	table.SetAutoFormatHeaders(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("|")
	table.AppendBulk(params) // Add Bulk Data
	table.Render()

	doc.Chart.Values = buf.String()
	for _, row := range rows {
		if row[2] != "" &&
			row[2] != `""` &&
			row[2] != "{}" &&
			row[2] != "[]" &&
			row[2] != "true" &&
			row[2] != "false" &&
			row[2] != "not-ca-cert" {
			doc.Chart.ValuesExample = fmt.Sprintf("%v=%v", row[0], row[2])
			break
		}
	}

	tplReadme, err := ioutil.ReadFile(*tplFile)
	if err != nil {
		if os.IsNotExist(err) {
			tplReadme, err = templates.Asset("readme.tpl")
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	tmpl, err := template.New("readme").Parse(string(tplReadme))
	if err != nil {
		panic(err)
	}

	var output *os.File

	if len(*outFile) == 0 {
		output = os.Stdout
	} else {
		output, err = os.OpenFile(*outFile, os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	err = tmpl.Execute(output, doc)
	if err != nil {
		panic(err)
	}
}
