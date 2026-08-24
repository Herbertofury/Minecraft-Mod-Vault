package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
)

type testGridJUnitSuite struct {
	XMLName  xml.Name            `xml:"testsuite"`
	Name     string              `xml:"name,attr"`
	Tests    int                 `xml:"tests,attr"`
	Failures int                 `xml:"failures,attr"`
	Time     string              `xml:"time,attr"`
	Cases    []testGridJUnitCase `xml:"testcase"`
}

type testGridJUnitCase struct {
	Name    string                `xml:"name,attr"`
	Class   string                `xml:"classname,attr"`
	Time    string                `xml:"time,attr"`
	Failure *testGridJUnitFailure `xml:"failure,omitempty"`
	Skipped *struct{}             `xml:"skipped,omitempty"`
}

type testGridJUnitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

func writeTestGridJUnit(path string, run TestGridRun) error {
	suite := testGridJUnitSuite{Name: run.Name, Tests: len(run.Steps), Time: fmt.Sprintf("%.3f", float64(run.DurationMS)/1000)}
	for _, step := range run.Steps {
		item := testGridJUnitCase{Name: step.Name, Class: "minecraft-mod-vault.testgrid." + step.Type, Time: fmt.Sprintf("%.3f", float64(step.DurationMS)/1000)}
		if step.Status == "failed" {
			suite.Failures++
			item.Failure = &testGridJUnitFailure{Message: step.Message, Body: step.Message}
		} else if step.Status == "skipped" {
			item.Skipped = &struct{}{}
		}
		suite.Cases = append(suite.Cases, item)
	}
	body, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}
	body = append([]byte(xml.Header), body...)
	body = append(body, '\n')
	return os.WriteFile(path, body, 0o644)
}

var testGridHTMLTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Name}} — TestGrid</title><style>
body{font:15px system-ui,sans-serif;max-width:1080px;margin:32px auto;padding:0 20px;background:#11141b;color:#edf2f7}h1{margin-bottom:4px}.meta{color:#aeb9c6}.passed{color:#5ee09a}.failed,.canceled{color:#ff7b8b}.skipped{color:#ffd36a}table{width:100%;border-collapse:collapse;margin-top:22px}th,td{text-align:left;padding:10px;border-bottom:1px solid #313845}code{color:#8de7ff}a{color:#8de7ff}</style></head>
<body><h1>{{.Name}}</h1><p class="meta">Run <code>{{.ID}}</code> · {{.Edition}} {{.GameVersion}} {{.Loader}}</p>
<h2 class="{{.Status}}">{{.Status}}</h2><p>{{.Error}}</p>
<table><thead><tr><th>Step</th><th>Type</th><th>Status</th><th>Time</th><th>Evidence</th></tr></thead><tbody>
{{range .Steps}}<tr><td>{{.Name}}</td><td><code>{{.Type}}</code></td><td class="{{.Status}}">{{.Status}}</td><td>{{.DurationMS}} ms</td><td>{{.Message}}</td></tr>{{end}}
</tbody></table><p class="meta">Process exit {{.Process.ExitCode}} · {{.DurationMS}} ms total</p></body></html>`))

func writeTestGridHTML(path string, run TestGridRun) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	executeErr := testGridHTMLTemplate.Execute(file, run)
	closeErr := file.Close()
	if executeErr != nil {
		return executeErr
	}
	return closeErr
}
