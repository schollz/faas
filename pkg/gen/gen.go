
// gen will generate new code
package gen 



func codeGeneration(packageName string, functionName string, inputParams []models.Param, outputParams []models.Param) (code string, err error) {

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"title": strings.Title,
	}

	const templateText = `
type Input struct {
	{{- range .InputParams }}
	{{title .Name }} {{.Type }} ` + "`" + `json:"{{.Name}}"` + "`" + `{{ end }}
}

var params Input
err = json.Unmarshal(b, &params)
{{range $index, $element := .OutputParams }}{{if $index}}, {{end}}{{$element.Name}}{{ end }} := {{.FunctionName}}(
	{{- range .InputParams}}
	params.{{title .Name }},{{end }}
)

// create json
fullJson = "{"
{{range $index, $element := .OutputParams }}
{{if $index}}fullJson += ","{{end}}
b, err = json.Marshal({{.Name}})
if err != nil {
	log.Error(err)
	return
}
fullJson +=  ` + "`" + `"{{.Name}}": ` + "`" + ` + string(b)
{{end}}
fullJson += "}"
`

	type TemplateStruct struct {
		FunctionName string
		InputParams  []models.Param
		OutputParams []models.Param
	}

	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(templateText)
	if err != nil {
		log.Error(err)
		return
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, TemplateStruct{
		packageName + "." + functionName, inputParams, outputParams,
	})
	if err != nil {
		log.Error(err)
		return
	}

	codeBytes, err := format.Source(tpl.Bytes())
	code = string(codeBytes)
	return
}
