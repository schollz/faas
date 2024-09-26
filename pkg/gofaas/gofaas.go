package gofaas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/schollz/faas/pkg/utils"
	log "github.com/schollz/logger"
)

type Param struct {
	Name string
	Type string
}

type CodeGen struct {
	ImportPath   string
	PackageName  string
	FunctionName string
	InputParams  []Param
	OutputParams []Param
}

var ErrorFunctionNotFound = errors.New("function not found")

func BuildContainer(importPathOrURL string, functionName string, containerName string) (err error) {
	// create a temp directory
	tempdir, err := ioutil.TempDir("", "build")
	if err != nil {
		log.Error(err)
		return
	}
	if log.GetLevel() != "debug" {
		defer os.RemoveAll(tempdir)
	}
	log.Debugf("working in %s", tempdir)

	if strings.HasPrefix(importPathOrURL, "http") {
		err = GenerateContainerFromURL(importPathOrURL, functionName, tempdir)
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		err = GenerateContainerFromImportPath(importPathOrURL, functionName, tempdir)
		if err != nil {
			log.Error(err)
			return
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}

	imagesPath := path.Join(cwd, "images")

	defer os.Chdir(cwd)
	absCwd, _ := filepath.Abs(tempdir)
	log.Tracef("cd into %s", absCwd)
	os.Chdir(tempdir)

	stdout, stderr, err := utils.RunCommand(fmt.Sprintf("docker build -t %s .", containerName))
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if stderr != "" {
		err = fmt.Errorf("%s\n%s", stdout, stderr)
		return
	}

	stdout, stderr, err = utils.RunCommand(fmt.Sprintf("docker save %s -o %s", containerName, path.Join(imagesPath, containerName+".tar")))
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if stderr != "" {
		err = fmt.Errorf("%s\n%s", stdout, stderr)
		return
	}

	stdout, stderr, err = utils.RunCommand(fmt.Sprintf("gzip %s", path.Join(imagesPath, containerName+".tar")))
	log.Debugf("stdout: [%s]", stdout)
	log.Debugf("stderr: [%s]", stderr)
	if stderr != "" {
		err = fmt.Errorf("%s\n%s", stdout, stderr)
		return
	}

	return
}

func GenerateContainerFromURL(urlString string, functionName string, tempdir string) (err error) {
	log.Debugf("building %s into %s", urlString, tempdir)
	resp, err := http.Get(urlString)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(path.Join(tempdir, "1.go"))
	if err != nil {
		log.Error(err)
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)

	// build the template file
	b, err := ioutil.ReadFile("template/main.go")
	if err != nil {
		log.Error(err)
		return
	}
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(string(b))
	if err != nil {
		log.Error(err)
		return
	}

	packageName, inputParams, outputParams, err := FindFunctionInFile(path.Join(tempdir, "1.go"), functionName)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("packageName: %+v", packageName)
	log.Debugf("inputParams: %+v", inputParams)
	log.Debugf("outputParams: %+v", outputParams)

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, CodeGen{
		ImportPath:   "",
		PackageName:  "",
		FunctionName: functionName,
		InputParams:  inputParams,
		OutputParams: outputParams,
	})
	if err != nil {
		log.Error(err)
		return
	}

	b, err = ioutil.ReadFile(path.Join(tempdir, "1.go"))
	if err != nil {
		log.Error(err)
		return
	}
	b = bytes.Replace(b, []byte("package "+packageName), []byte("package main"), 1)
	err = ioutil.WriteFile(path.Join(tempdir, "1.go"), b, 0644)
	if err != nil {
		log.Error(err)
		return
	}
	code := tpl.String()
	log.Debugf("code: %s", code)

	err = ioutil.WriteFile(path.Join(tempdir, "main.go"), []byte(code), 0644)
	if err != nil {
		log.Error(err)
		return
	}

	err = ioutil.WriteFile(path.Join(tempdir, "Dockerfile"), []byte(Dockerfile), 0644)
	if err != nil {
		log.Error(err)
		return
	}

	err = ioutil.WriteFile(path.Join(tempdir, "go.mod"), []byte(`module main`), 0644)
	if err != nil {
		log.Error(err)
		return
	}
	return
}

func GenerateContainerFromImportPath(importPath string, functionName string, tempdir string) (err error) {
	log.Debugf("building %s into %s", importPath, tempdir)

	// build the template file
	b, err := ioutil.ReadFile("template/main.go")
	if err != nil {
		log.Error(err)
		return
	}
	funcMap := template.FuncMap{
		"title": strings.Title,
	}
	tmpl, err := template.New("titleTest").Funcs(funcMap).Parse(string(b))
	if err != nil {
		log.Error(err)
		return
	}

	packageName, inputParams, outputParams, err := FindFunctionInImportPath(importPath, functionName)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("packageName: %+v", packageName)
	log.Debugf("inputParams: %+v", inputParams)
	log.Debugf("outputParams: %+v", outputParams)
	log.Debugf("importPath: %+v", importPath)
	log.Debugf("functionName: %+v", functionName)
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, CodeGen{
		ImportPath:   importPath,
		PackageName:  packageName,
		FunctionName: functionName,
		InputParams:  inputParams,
		OutputParams: outputParams,
	})
	if err != nil {
		log.Error(err)
		return
	}

	code := tpl.String()
	log.Debugf("code: %s", code)
	err = ioutil.WriteFile(path.Join(tempdir, "main.go"), []byte(code), 0644)
	if err != nil {
		log.Error(err)
		return
	}

	err = ioutil.WriteFile(path.Join(tempdir, "Dockerfile"), []byte(Dockerfile), 0644)
	if err != nil {
		log.Error(err)
		return
	}

	err = ioutil.WriteFile(path.Join(tempdir, "go.mod"), []byte(`module main`), 0644)
	if err != nil {
		log.Error(err)
		return
	}

	return
}

// FindFunctionInImportPath takes an import path and a function name and returns
// all the nessecary components to generate the file
func FindFunctionInImportPath(importPath string, functionName string) (packageName string, inputParams []Param, outputParams []Param, err error) {
	// create a temp directory
	tempdir, err := ioutil.TempDir("", "parser")
	if err != nil {
		log.Error(err)
		return
	}
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

	// loop through the files to find the function
	for _, fname := range goFiles {
		packageName, inputParams, outputParams, err = FindFunctionInFile(fname, functionName)
		if err == nil {
			for i := range inputParams {
				inputParams[i].Type = UpdateTypeWithPackage(packageName, inputParams[i].Type)
			}
			for i := range inputParams {
				outputParams[i].Type = UpdateTypeWithPackage(packageName, outputParams[i].Type)
			}
			return
		}
	}
	return
}

func FindFunctionInFile(fname string, functionName string) (packageName string, inputParams []Param, outputParams []Param, err error) {
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
			log.Debugf("found funnction: %s", fd.Name.Name)
			if fd.Name.Name != functionName {
				return true
			}
			// found function
			err = nil
			inputParams = make([]Param, len(fd.Type.Params.List))
			for i, param := range fd.Type.Params.List {
				if len(param.Names) == 0 {
					continue
				}
				inputParams[i] = Param{
					param.Names[0].Name,
					src[param.Type.Pos()-offset : param.Type.End()-offset],
				}
			}
			outputParams = make([]Param, len(fd.Type.Results.List))
			for i, param := range fd.Type.Results.List {
				if len(param.Names) == 0 {
					continue
				}
				outputParams[i] = Param{
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

// ParseFunctionString takes input like ParseFunctionString("x","y","z"),`run(1,"hello",[1.2,2.1])`)
// and returns []byte(`{"x":1,"y":"hello","z":[1.2,2.1]}`) which can later be used for unmarshalling
func ParseFunctionString(paramNames []string, functionString string) (functionName string, jsonBytes []byte, err error) {
	if !strings.Contains(functionString, "(") {
		err = fmt.Errorf("must contain ()")
		log.Error(err)
		return
	}
	// add brackets
	foo := strings.SplitN(functionString, "(", 2)
	functionName = strings.TrimSpace(foo[0])
	functionString = strings.TrimSpace(foo[1])
	functionString = "[" + functionString[:len(functionString)-1] + "]"

	var values []interface{}
	err = json.Unmarshal([]byte(functionString), &values)
	if err != nil {
		log.Error(err)
		return
	}

	if len(values) != len(paramNames) {
		err = fmt.Errorf("number of values and param names not equal")
		log.Error(err)
		return
	}

	// build JSON string
	jsonString := ""
	for i, value := range values {
		var valueByte []byte
		valueByte, err = json.Marshal(value)
		if err != nil {
			log.Error(err)
			return
		}
		jsonString += `"` + paramNames[i] + `": ` + string(valueByte)
		if i < len(values)-1 {
			jsonString += ", "
		}
	}

	jsonString = "{" + jsonString + "}"
	jsonBytes = []byte(jsonString)
	return
}

var types = []string{"error", "string", "bool", "byte", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64", "int", "uint", "uintptr", "float32", "float64", "complex64", "complex128"}

func UpdateTypeWithPackage(packageName string, typeString string) (newTypeString string) {
	typeString = strings.TrimPrefix(typeString, "...")
	newTypeString = typeString
	if strings.Contains(typeString, ".") {
		// don't handle the other types yet
		return
	}
	isarray := typeString[:2] == "[]"
	if isarray {
		typeString = typeString[2:]
	}
	ispointer := string(typeString[0]) == string("*")
	if ispointer {
		typeString = typeString[1:]
	}
	isnormal := false
	for _, t := range types {
		if typeString == t {
			isnormal = true
			break
		}
	}
	if isnormal {
		return newTypeString
	}
	newTypeString = ""
	if isarray {
		newTypeString += "[]"
	}
	if ispointer {
		newTypeString += "*"
	}
	newTypeString += packageName + "." + typeString
	return
}

const Dockerfile = `
##################################
# 1. Build in a Go-based image   #
###################################
FROM golang as builder
RUN apk add git
WORKDIR /go/main
COPY . .
ENV GO111MODULE=on
RUN go build -v

###################################
# 2. Copy into a clean image     #
###################################
FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/main/main /main
EXPOSE 8080
ENTRYPOINT ["/main"]
# any flags here, for example use the data folder
CMD ["--debug"] 
`
