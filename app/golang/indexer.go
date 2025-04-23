package golang

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brandtg/rtfm/app/common"
)

func Index() error {
	slog.Info("Indexing Go modules...")
	// Connect to the database
	db, err := common.OpenDB()
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	// Find GOPATH
	gopath, err := findGoPath()
	if err != nil {
		return fmt.Errorf("error resolving GOPATH: %w", err)
	}
	// Find modules in the GOPATH
	modules, err := findModules(gopath)
	if err != nil {
		return fmt.Errorf("error finding modules: %w", err)
	}
	for _, module := range modules {
		// Find code files in the module
		codeFiles, err := findCodeFiles(module)
		if err != nil {
			slog.Error("Error finding code files", "module", module, "error", err)
			continue
		}
		// Index the documents
		err = common.IndexDocuments(db, codeFiles)
		if err != nil {
			return fmt.Errorf("error indexing documents: %w", err)
		}
	}
	return nil
}

func findGoPath() (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = os.ExpandEnv("$HOME/go") // Default GOPATH
	}
	if _, err := os.Stat(gopath); os.IsNotExist(err) {
		return "", fmt.Errorf("GOPATH directory does not exist: %s", gopath)
	}
	return gopath, nil
}

func findModules(gopath string) ([]string, error) {
	acc := make([]string, 0)
	err := filepath.Walk(gopath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Base(path) == "go.mod" {
			acc = append(acc, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", gopath, err)
	}
	return acc, nil
}

// var moduleNameRegex = regexp.MustCompile(`^module\s+([^\s]+)`)
var moduleNameRegex = regexp.MustCompile(`(?m)^module\s+([^\s]+)`)

func findModuleName(path string) (string, error) {
	// Read the go.mod file
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading go.mod file: %w", err)
	}
	// Parse the module name
	matches := moduleNameRegex.FindSubmatch(bytes)
	if len(matches) < 2 {
		return "", fmt.Errorf("error parsing module name from go.mod file: %s", path)
	}
	moduleName := string(matches[1])
	// Clean the quotes
	moduleName = strings.Trim(moduleName, "\"")
	return moduleName, nil
}

func findCodeFiles(module string) ([]*common.SearchDocument, error) {
	// Find the module name
	moduleName, err := findModuleName(module)
	if err != nil {
		return nil, fmt.Errorf("error finding module name: %w", err)
	}
	// Find the code files
	moduleDir := filepath.Dir(module)
	codeFiles := make([]string, 0)
	err = filepath.Walk(
		moduleDir,
		func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() &&
				filepath.Ext(path) == ".go" &&
				!strings.Contains(filepath.Base(path), "_test.go") {
				codeFiles = append(codeFiles, path)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("error walking the path %s: %w", module, err)
	}
	// Construct search documents
	docs := make([]*common.SearchDocument, len(codeFiles))
	for i, codeFile := range codeFiles {
		name := strings.Replace(codeFile, moduleDir, moduleName, 1)
		docs[i] = &common.SearchDocument{
			Language: common.Go,
			Name:     name,
			Path:     codeFile,
		}
	}
	return docs, nil
}
