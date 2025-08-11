package utils

import (
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func WriteFile(fset *token.FileSet, astFile *ast.File) error {
	f, err := os.OpenFile(fset.File(astFile.FileStart).Name(), os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := format.Node(f, fset, astFile); err != nil {
		return err
	}

	return nil
}

func FormatFile(fileName string) {
	cmd := exec.Command("gofumpt", "-w", fileName)
	if err := cmd.Run(); err != nil {
		logger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
		logger.Printf("Error running gofumpt: %s", err.Error())
	}
}

func FindDir(fullPath, dir string) (string, error) {
	currentDir := filepath.Dir(fullPath)

	for currentDir != "/" && filepath.Base(currentDir) != dir {
		currentDir = filepath.Dir(currentDir)
	}

	if filepath.Base(currentDir) != dir {
		return "", os.ErrNotExist
	}

	return currentDir, nil
}
