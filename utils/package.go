package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type PkgNames struct {
	Plural   string
	Singular string
}

var CqMap = map[string]PkgNames{
	"Command": {
		Plural:   "commands",
		Singular: "command",
	},
	"Query": {
		Plural:   "queries",
		Singular: "query",
	},
}

func ExecutionPath() string {
	return filepath.Join(os.Getenv("PWD"), os.Getenv("GOFILE"))
}

func ImportPath(moduleName, moduleRoot, filename string) string {
	return moduleName + strings.TrimPrefix(filepath.Dir(filename), moduleRoot)
}

func ModuleInfo() (string, string, error) {
	dir := os.Getenv("PWD")
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			file, err := os.Open(modPath)
			if err != nil {
				return "", "", err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					return dir, moduleName, nil
				}
			}
			return dir, "", scanner.Err()
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", os.ErrNotExist
		}
		dir = parent
	}
}
