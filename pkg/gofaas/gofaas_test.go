package gofaas

import (
	"testing"

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

// func TestCodeGeneration(t *testing.T) {
// 	packageName := "parser"
// 	functionName := "FindFunction"
// 	inputParams := []Param{Param{Name: "gitURL", Type: "string"}, Param{Name: "functionName", Type: "string"}}
// 	outputParams := []Param{Param{Name: "structString", Type: "string"}, Param{Name: "err", Type: "error"}}
// 	code, err := codeGeneration(packageName, functionName, inputParams, outputParams)
// 	assert.Nil(t, err)
// 	codeGood := "\ntype Input struct {\n\tGitURL       string `json:\"gitURL\"`\n\tFunctionName string `json:\"functionName\"`\n}\n\ntype Output struct {\n\tStructString string `json:\"structString\"`\n\tErr          error  `json:\"err\"`\n}\n\nvar params Input\nvar result Output\nerr = json.Unmarshal(b, &params)\nresult.StructString, result.Err = parser.FindFunction(\n\tparams.GitURL,\n\tparams.FunctionName,\n)\n\n"
// 	assert.Equal(t, codeGood, code)
// 	fmt.Println(codeGood)
// 	fmt.Println(code)
// }
