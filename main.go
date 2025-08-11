package main

import (
	"go/parser"
	"go/token"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/generators/application"
	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/generators/cq"
	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/generators/handler"
	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/utils"
	//. "github.com/dave/jennifer/jen"
)

func main() {
	log.SetOutput(os.Stdout)
	fullPath := utils.ExecutionPath()

	moduleRoot, moduleName, err := utils.ModuleInfo()
	if err != nil {
		logFatalf("Error finding module info: %v", err)
	}
	pkg := os.Getenv("GOPACKAGE")
	importPath := utils.ImportPath(moduleName, moduleRoot, fullPath)

	_, fileName := filepath.Split(fullPath)
	cqType := strings.Split(fileName, ".")[0]
	cqType = strings.ToUpper(cqType[:1]) + cqType[1:]
	if cqType != "Query" && cqType != "Command" {
		logFatalf("filename must be either 'query' or 'command', got: %s", cqType)
	}

	cmd := exec.Command("gofumpt", "-h")
	if err := cmd.Run(); err != nil {
		logFatalf("please install gofumpt to format the code: `go install mvdan.cc/gofumpt@latest`")
	}

	applicationDir, err := utils.FindDir(utils.ExecutionPath(), "application")
	if err != nil {
		logFatalf("error finding applicaiton dir: %s", err.Error())
	}

	fset := token.NewFileSet()
	parsedDir, err := parser.ParseDir(fset, applicationDir, nil, parser.ParseComments)
	if err != nil {
		logFatalf("error parsing directory: %s", err.Error())
	}

	cqGenerator := cq.NewCQGenerator(pkg, cqType, importPath)
	if err := cqGenerator.Generate(fset, maps.Values(parsedDir["application"].Files)); err != nil {
		logFatalf("error generating cq: %s", err.Error())
	}
	applicationGenerator := application.NewApplicationGenerator(pkg, cqType, importPath)
	if err := applicationGenerator.Generate(fset, maps.Values(parsedDir["application"].Files)); err != nil {
		logFatalf("error generating application: %s", err.Error())
	}
	handlerGenerator := handler.New(pkg, cqType, filepath.Dir(fullPath))
	if err := handlerGenerator.Generate(); err != nil {
		logFatalf("error generating handler: %s", err.Error())
	}
}

func logFatalf(format string, args ...interface{}) {
	logger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	logger.Fatalf(format, args...)
}
