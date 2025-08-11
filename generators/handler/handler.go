package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/PickDough/cqrs-generator/utils"
	"github.com/dave/jennifer/jen"
)

type HandlerGenerator struct {
	CQPackage string
	CQType    string
	DirPath   string
}

func New(CQPackage string, CQType string, dirPath string) *HandlerGenerator {
	return &HandlerGenerator{CQPackage: CQPackage, CQType: CQType, DirPath: dirPath}
}

func (h *HandlerGenerator) Generate() error {
	filePath := h.DirPath + "/handler.go"
	if _, err := os.Stat(filePath); err == nil || !os.IsNotExist(err) {
		log.Printf("File %s already exists, skipping generation\n", filePath)
		return nil
	}

	f := h.generateFile()

	err := h.writeFile(filePath, fmt.Sprintf("%#v", f))
	if err != nil {
		return err
	}
	utils.FormatFile(filePath)

	return nil
}

func (h *HandlerGenerator) writeFile(filePath string, src string) error {
	fh, err := os.Create(filePath)
	if err != nil {
		return err
	}
	_, err = fh.WriteString(src)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", filePath, err)
	}

	defer func(fh *os.File) {
		_ = fh.Close()
	}(fh)

	return nil
}

func (h *HandlerGenerator) generateFile() *jen.File {
	f := jen.NewFile(h.CQPackage)
	f.Type().Id("handler").Struct()
	f.Func().Id("New").
		Params().
		Id("*handler").
		Block(jen.Return(jen.Id("&handler{}")))
	execFunc := f.Func().Params(jen.Id("h").Id("*handler")).
		Id("Execute").
		Params(jen.Id("ctx").Qual("context", "Context"), jen.Id(utils.CqMap[h.CQType].Singular).Id(h.CQType))
	if h.CQType == "Query" {
		execFunc.Params(jen.Id("Result"), jen.Id("error"))
	} else {
		execFunc.Id("error")
	}
	execFunc.Block(
		jen.Panic(jen.Lit("not implemented")).Line(),
	)

	return f
}
