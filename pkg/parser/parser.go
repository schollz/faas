package parser

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/faasss/pkg/models"
	"github.com/schollz/faasss/pkg/utils"
	log "github.com/schollz/logger"
)

var ErrorFunctionNotFound = errors.New("function not found")

func FindFunction(importPath string, functionName string) (structString string, err error) {
	// create a temp directory
	tempdir, err := ioutil.TempDir("", "parser")
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("cloning %s into %s", importPath, tempdir)
	defer os.RemoveAll(tempdir)

	// clone into temp directory
	stdout, stderr, err := utils.RunCommand(fmt.Sprintf("git clone --depth 1 https://%s %s", importPath, tempdir))
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if err != nil {
		log.Error(err)
		return
	}
	if strings.Contains(stderr, "fatal") {
		err = fmt.Errorf("%s", stderr)
		return
	}

	// find all go files
	goFiles := []string{}
	err = filepath.Walk(tempdir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				goFiles = append(goFiles, path)
			}
			return nil
		})
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("found %d go files", len(goFiles))

	// return the function
	return
}

func findFunctionInFile(fname string, functionName string) (packageName string, inputParams []models.Param, outputParams []models.Param, err error) {
	// read file
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Error(err)
		return
	}
	src := string(b)

	// create token set
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, src, 0)
	if err != nil {
		log.Error(err)
		return
	}
	offset := f.Pos()

	// look for function in file, default is not found
	err = ErrorFunctionNotFound
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.File); ok {
			packageName = fd.Name.Name
		}
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Name.Name != functionName {
				return true
			}
			// found function
			err = nil
			inputParams = make([]models.Param, len(fd.Type.Params.List))
			for i, param := range fd.Type.Params.List {
				inputParams[i] = models.Param{
					param.Names[0].Name,
					src[param.Type.Pos()-offset : param.Type.End()-offset],
				}
			}
			outputParams = make([]models.Param, len(fd.Type.Results.List))
			for i, param := range fd.Type.Results.List {
				outputParams[i] = models.Param{
					param.Names[0].Name,
					src[param.Type.Pos()-offset : param.Type.End()-offset],
				}
			}
		}
		return true
	})
	if packageName == "" {
		err = errors.New("no package name")
	}
	return
}

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
