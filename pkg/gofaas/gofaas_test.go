package gofaas

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"text/template"

	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel("trace")
}

func TestFindFunctionInFile(t *testing.T) {
	packageName, inputParams, outputParams, err := FindFunctionInFile("gofaas.go", "FindFunctionInFile")
	assert.Nil(t, err)
	assert.Equal(t, "gofaas", packageName)
	assert.Equal(t, []Param{Param{Name: "fname", Type: "string"}, Param{Name: "functionName", Type: "string"}}, inputParams)
	assert.Equal(t, []Param{Param{Name: "packageName", Type: "string"}, Param{Name: "inputParams", Type: "[]Param"}, Param{Name: "outputParams", Type: "[]Param"}, Param{Name: "err", Type: "error"}}, outputParams)
	_, _, _, err = FindFunctionInFile("gofaas.go", "DoesntExist")
	assert.NotNil(t, err)
}

func TestParseFunctionString(t *testing.T) {
	functionName, jsonBytes, err := ParseFunctionString([]string{"x", "y", "z"}, `run(1,"hello",[1.2,2.1])`)
	assert.Nil(t, err)
	assert.Equal(t, "run", functionName)
	assert.Equal(t, `{"x": 1, "y": "hello", "z": [1.2,2.1]}`, string(jsonBytes))
}

func TestFindFunctionInImportPath(t *testing.T) {
	packageName, inputParams, outputParams, err := FindFunctionInImportPath("github.com/schollz/ingredients", "NewFromURL")
	assert.Nil(t, err)
	assert.Equal(t, "ingredients", packageName)
	assert.Equal(t, []Param{Param{Name: "url", Type: "string"}}, inputParams)
	assert.Equal(t, []Param{Param{Name: "r", Type: "*ingredients.Recipe"}, Param{Name: "err", Type: "error"}}, outputParams)
}

func TestUpdateTypeWithPackage(t *testing.T) {
	assert.Equal(t, "[]*ingredients.Recipe", UpdateTypeWithPackage("ingredients", "[]*Recipe"))
	assert.Equal(t, "int32", UpdateTypeWithPackage("ingredients", "int32"))
	assert.Equal(t, "models.Param", UpdateTypeWithPackage("ingredients", "models.Param"))
}

func TestCodeGeneration(t *testing.T) {
	type CodeGen struct {
		ImportPath   string
		PackageName  string
		FunctionName string
		InputParams  []Param
		OutputParams []Param
	}
	b, _ := ioutil.ReadFile("template/main.go")
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(string(b))
	if err != nil {
		log.Error(err)
		return
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, CodeGen{
		ImportPath:   "github.com/schollz/ingredients",
		PackageName:  "ingredients",
		FunctionName: "NewFromURL",
		InputParams:  []Param{Param{Name: "url", Type: "string"}},
		OutputParams: []Param{Param{Name: "r", Type: "*ingredients.Recipe"}, Param{Name: "err", Type: "error"}},
	})
	assert.Nil(t, err)

	code := tpl.String()
	fmt.Println(code)
	ioutil.WriteFile("test/1.go", []byte(code), 0644)
	assert.Nil(t, nil)
}
