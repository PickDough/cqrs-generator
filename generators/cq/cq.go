package cq

import (
	"fmt"
	"go/ast"
	"go/token"
	"iter"
	"log"
	"strings"

	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/utils"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

type cqGenerator struct {
	CQPackage  string
	CQType     string
	ImportPath string
}

func NewCQGenerator(cqPackage string, cqType string, importPath string) *cqGenerator {
	return &cqGenerator{
		CQPackage:  cqPackage,
		CQType:     cqType,
		ImportPath: importPath,
	}
}

func (cq *cqGenerator) Generate(fset *token.FileSet, applicationPackage iter.Seq[*ast.File]) error {
	astFile := cq.findFileToInsert(applicationPackage)
	if astFile == nil {
		return fmt.Errorf("couldn't find %s location inside application directory", cq.CQType)
	}

	for _, vv := range astutil.Imports(fset, astFile) {
		for _, impor := range vv {
			if impor.Path.Value == fmt.Sprintf("\"%s\"", cq.ImportPath) {
				log.Printf("Import %s already exists in %s, skipping generation", cq.ImportPath, fset.File(astFile.FileStart).Name())
				return nil
			}
		}
	}

	cq.insertField(astFile)
	cq.insertInterface(astFile)
	astutil.AddImport(fset, astFile, cq.ImportPath)

	utils.WriteFile(fset, astFile)
	utils.FormatFile(fset.File(astFile.FileStart).Name())

	return nil
}

func (cq *cqGenerator) findFileToInsert(applicationPackage iter.Seq[*ast.File]) *ast.File {
	for file := range applicationPackage {
		inspec := inspector.New([]*ast.File{file})
		for ts := range inspec.Root().Preorder((*ast.TypeSpec)(nil)) {
			typeSpec := ts.Node().(*ast.TypeSpec)
			if strings.Contains(strings.ToLower(typeSpec.Name.Name), utils.SubStringCQ[cq.CQType].Plural) {
				return file
			}
		}
	}

	return nil
}

func (cq *cqGenerator) insertField(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || !strings.Contains(strings.ToLower(typeSpec.Name.Name), utils.SubStringCQ[cq.CQType].Plural) {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return false
		}

		newField := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(cq.CQPackage)},
			Type:  ast.NewIdent(cq.CQPackage + cq.CQType + "Handler"),
		}

		structType.Fields.List = append(structType.Fields.List, newField)

		return false
	})
}

func (cq *cqGenerator) insertInterface(node *ast.File) {
	// Build the interface type directly
	methodFields := []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent("Execute")},
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{ast.NewIdent("ctx")},
							Type:  ast.NewIdent("context.Context"),
						},
						{
							Names: []*ast.Ident{ast.NewIdent(utils.SubStringCQ[cq.CQType].Singular)},
							Type:  ast.NewIdent(fmt.Sprintf("%s.%s", cq.CQPackage, cq.CQType)),
						},
					},
				},
				Results: &ast.FieldList{
					List: func() []*ast.Field {
						if cq.CQType == "Query" {
							return []*ast.Field{
								{Type: ast.NewIdent(fmt.Sprintf("%s.Result", cq.CQPackage))},
								{Type: ast.NewIdent("error")},
							}
						}
						return []*ast.Field{
							{Type: ast.NewIdent("error")},
						}
					}(),
				},
			},
		},
	}

	interfaceType := &ast.InterfaceType{
		Methods: &ast.FieldList{
			List: methodFields,
		},
	}

	interfaceSpec := &ast.TypeSpec{
		Name: ast.NewIdent(fmt.Sprintf("%s%sHandler", cq.CQPackage, cq.CQType)),
		Type: interfaceType,
	}

	// Insert into the first type declaration group, or create a new one
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			genDecl.Specs = append(genDecl.Specs, interfaceSpec)
			return
		}
	}

	typeDecl := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{interfaceSpec},
	}
	node.Decls = append(node.Decls, typeDecl)
}
