package gen

import (
	"fmt"
	"testing"

	"github.com/schollz/faas/pkg/models"
	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel("debug")
}

func TestCodeGeneration(t *testing.T) {
	packageName := "parser"
	functionName := "FindFunction"
	inputParams := []models.Param{models.Param{Name: "gitURL", Type: "string"}, models.Param{Name: "functionName", Type: "string"}}
	outputParams := []models.Param{models.Param{Name: "structString", Type: "string"}, models.Param{Name: "err", Type: "error"}}
	code, err := codeGeneration(packageName, functionName, inputParams, outputParams)
	assert.Nil(t, err)
	codeGood := "\ntype Input struct {\n\tGitURL       string `json:\"gitURL\"`\n\tFunctionName string `json:\"functionName\"`\n}\n\ntype Output struct {\n\tStructString string `json:\"structString\"`\n\tErr          error  `json:\"err\"`\n}\n\nvar params Input\nvar result Output\nerr = json.Unmarshal(b, &params)\nresult.StructString, result.Err = parser.FindFunction(\n\tparams.GitURL,\n\tparams.FunctionName,\n)\n\n"
	assert.Equal(t, codeGood, code)
	fmt.Println(codeGood)
	fmt.Println(code)
}
