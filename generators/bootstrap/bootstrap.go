package bootstrap

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"github.com/PickDough/cqrs-generator/utils"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

type BootstrapGenerator struct {
	CQPackage  string
	CQType     string
	ImportPath string
}

func NewBootstrapGenerator(cqPackage, cqType, importPath string) *BootstrapGenerator {
	return &BootstrapGenerator{
		CQPackage:  cqPackage,
		CQType:     cqType,
		ImportPath: importPath,
	}
}

func (b *BootstrapGenerator) Generate(fset *token.FileSet, astFile *ast.File) error {
	for _, vv := range astutil.Imports(fset, astFile) {
		for _, impor := range vv {
			if impor.Path.Value == fmt.Sprintf("\"%s\"", b.ImportPath) {
				log.Printf("Import %s already exists in %s, skipping generation", b.ImportPath, fset.File(astFile.FileStart).Name())
				return nil
			}
		}
	}

	inspec := inspector.New([]*ast.File{astFile})
	b.insertArgument(inspec)
	astutil.AddNamedImport(fset, astFile, b.CQPackage+b.CQType+"Handler", b.ImportPath)

	err := utils.WriteFile(fset, astFile)
	if err != nil {
		return err
	}
	utils.FormatFile(fset.File(astFile.FileStart).Name())

	return nil
}

func (b *BootstrapGenerator) insertArgument(inspec *inspector.Inspector) {
	for expr := range inspec.Root().Preorder((*ast.CallExpr)(nil)) {
		if ce, ok := expr.Node().(*ast.CallExpr); ok {
			if fun, ok := ce.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "application" && fun.Sel.Name == "New" {
					newCe := &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name:    b.CQPackage + b.CQType + "Handler",
								NamePos: ce.Rparen + 1,
							},
							Sel: ast.NewIdent("New"),
						},
					}
					ce.Args = append(ce.Args, newCe)
				}
			}
		}
	}
}
