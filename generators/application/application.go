package application

import (
	"fmt"
	"go/ast"
	"go/token"
	"iter"
	"log"

	"gitlab.com/social-tech/teams/bond-team/cqrs-generator/utils"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

type ApplicationGenerator struct {
	CQPackage  string
	CQType     string
	ImportPath string
}

func NewApplicationGenerator(cqPackage string, cqType string, importPath string) *ApplicationGenerator {
	return &ApplicationGenerator{
		CQPackage:  cqPackage,
		CQType:     cqType,
		ImportPath: importPath,
	}
}

func (a *ApplicationGenerator) Generate(fset *token.FileSet, applicationPackage iter.Seq[*ast.File]) error {
	astFile := a.findFileToInsert(applicationPackage)
	if astFile == nil {
		return fmt.Errorf("couldn't find %s location inside application directory", a.CQType)
	}

	inspec := inspector.New([]*ast.File{astFile})
	if a.hasParameter(inspec) {
		log.Printf("Parameter %s already exists in %s, skipping generation", a.CQPackage, fset.File(astFile.FileStart).Name())
		return nil
	}

	astutil.AddImport(fset, astFile, a.ImportPath)
	a.insertParameter(inspec)
	a.insertField(inspec)
	err := a.insertFunc(inspec, astFile)
	if err != nil {
		return err
	}

	err = utils.WriteFile(fset, astFile)
	if err != nil {
		return err
	}
	utils.FormatFile(fset.File(astFile.FileStart).Name())

	return nil
}

func (a *ApplicationGenerator) insertParameter(inspec *inspector.Inspector) {
	for fd := range inspec.Root().Preorder((*ast.FuncDecl)(nil)) {
		funcDecl := fd.Node().(*ast.FuncDecl)
		if funcDecl.Name.Name != "New" {
			continue
		}

		param := &ast.Field{
			Names: []*ast.Ident{
				{
					NamePos: funcDecl.Type.Params.Closing,
					Name:    fmt.Sprintf("%s%sHandler", a.CQPackage, a.CQType),
				},
			},
			Type: ast.NewIdent(fmt.Sprintf("%s%sHandler", a.CQPackage, a.CQType)),
		}

		funcDecl.Type.Params.List = append(funcDecl.Type.Params.List, param)
		funcDecl.Type.Params.Closing += 1

		return
	}
}

func (a *ApplicationGenerator) insertField(inspec *inspector.Inspector) {
	for fd := range inspec.Root().Preorder((*ast.CompositeLit)(nil)) {
		compositeLit := fd.Node().(*ast.CompositeLit)
		if ident, ok := compositeLit.Type.(*ast.Ident); ok && ident.Name == utils.CqMap[a.CQType].Plural {
			newField := &ast.KeyValueExpr{
				Key: &ast.Ident{
					NamePos: compositeLit.Rbrace,
					Name:    a.CQPackage,
				},
				Value: ast.NewIdent(fmt.Sprintf("%s%sHandler", a.CQPackage, a.CQType)),
			}

			compositeLit.Elts = append(compositeLit.Elts, newField)
			return
		}
	}
}

func (a *ApplicationGenerator) insertFunc(inspec *inspector.Inspector, astFile *ast.File) error {
	newFunc := a.findNewFunc(inspec)
	if newFunc == nil {
		return fmt.Errorf("couldn't find New function in %s", astFile.Name.Name)
	}

	recvType := newFunc.Type.Results.List[0].Type
	funcDecl := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						ast.NewIdent("app"),
					},
					Type: recvType,
				},
			},
		},
		Name: ast.NewIdent(utils.FirstLetterToUpper(a.CQPackage)),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							ast.NewIdent(fmt.Sprintf("ctx")),
						},
						Type: ast.NewIdent("context.Context"),
					},
					{
						Names: []*ast.Ident{
							ast.NewIdent(utils.CqMap[a.CQType].Singular),
						},
						Type: ast.NewIdent(fmt.Sprintf("%s.%s", a.CQPackage, a.CQType)),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: ast.NewIdent(fmt.Sprintf("app.%s.%s.Execute", utils.CqMap[a.CQType].Plural, a.CQPackage)),
							Args: []ast.Expr{
								ast.NewIdent("ctx"),
								ast.NewIdent(utils.CqMap[a.CQType].Singular),
							},
						},
					},
				},
			},
		},
	}
	if a.CQType == "Query" {
		funcDecl.Type.Results.List = append(
			[]*ast.Field{
				{
					Type: ast.NewIdent(fmt.Sprintf("%s.Result", a.CQPackage)),
				},
			},
			funcDecl.Type.Results.List...)

	}
	astFile.Decls = append(astFile.Decls, funcDecl)

	return nil
}

func (a *ApplicationGenerator) findFileToInsert(applicationPackage iter.Seq[*ast.File]) *ast.File {
	for file := range applicationPackage {
		inspec := inspector.New([]*ast.File{file})

		newFunc := a.findNewFunc(inspec)
		if newFunc != nil {
			return file
		}
	}

	return nil
}

func (a *ApplicationGenerator) findNewFunc(inspec *inspector.Inspector) *ast.FuncDecl {
	for fd := range inspec.Root().Preorder((*ast.FuncDecl)(nil)) {
		funcDecl := fd.Node().(*ast.FuncDecl)
		if funcDecl.Name.Name != "New" {
			continue
		}

		return funcDecl
	}

	return nil
}

func (a *ApplicationGenerator) hasParameter(inspec *inspector.Inspector) bool {
	for fd := range inspec.Root().Preorder((*ast.FuncDecl)(nil)) {
		funcDecl := fd.Node().(*ast.FuncDecl)
		if funcDecl.Name.Name != "New" {
			continue
		}
		for _, param := range funcDecl.Type.Params.List {
			for _, name := range param.Names {
				if name.Name == fmt.Sprintf("%s%sHandler", a.CQPackage, a.CQType) {
					return true
				}
			}
		}
	}

	return false
}
