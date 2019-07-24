package parser

import (
	"fmt"
	"testing"

	"github.com/schollz/faas/pkg/models"
	log "github.com/schollz/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel("trace")
}

func TestParser(t *testing.T) {
	structString, err := FindFunction("github.com/schollz/ingredients", "NewFromString")
	fmt.Println(structString)
	assert.Nil(t, err)
}

func TestFindFunction(t *testing.T) {
	packageName, inputParams, outputParams, err := findFunctionInFile("parser.go", "FindFunction")
	assert.Nil(t, err)
	assert.Equal(t, "parser", packageName)
	assert.Equal(t, []models.Param{models.Param{Name: "importPath", Type: "string"}, models.Param{Name: "functionName", Type: "string"}}, inputParams)
	assert.Equal(t, []models.Param{models.Param{Name: "structString", Type: "string"}, models.Param{Name: "err", Type: "error"}}, outputParams)
	_, _, _, err = findFunctionInFile("parser.go", "DoesntExist")
	assert.NotNil(t, err)
}
