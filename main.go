package main

import (
	"go/parser"
	"go/token"
	"log"
	"maps"
	"os"
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
		log.Fatalf("Error finding module info: %v", err)
	}
	pkg := os.Getenv("GOPACKAGE")
	importPath := utils.ImportPath(moduleName, moduleRoot, fullPath)

	_, fileName := filepath.Split(fullPath)
	cqType := strings.Split(fileName, ".")[0]
	cqType = strings.ToUpper(cqType[:1]) + cqType[1:]

	applicationDir, err := utils.FindDir(utils.ExecutionPath(), "application")
	if err != nil {
		log.Fatalf("error finding applicaiton dir: %s", err.Error())
	}

	fset := token.NewFileSet()
	parsedDir, err := parser.ParseDir(fset, applicationDir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("error parsing directory: %s", err.Error())
	}

	cqGenerator := cq.NewCQGenerator(pkg, cqType, importPath)
	if err := cqGenerator.Generate(fset, maps.Values(parsedDir["application"].Files)); err != nil {
		log.Fatalf("error generating cq: %s", err.Error())
	}
	applicationGenerator := application.NewApplicationGenerator(pkg, cqType, importPath)
	if err := applicationGenerator.Generate(fset, maps.Values(parsedDir["application"].Files)); err != nil {
		log.Fatalf("error generating application: %s", err.Error())
	}
	handlerGenerator := handler.New(pkg, cqType, filepath.Dir(fullPath))
	if err := handlerGenerator.Generate(); err != nil {
		log.Fatalf("error generating handler: %s", err.Error())
	}
}
