package python

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/brandtg/rtfm/app/common"
)

func Index() error {
	slog.Info("Indexing Python modules...")
	// Connect to the database
	db, err := common.OpenDB()
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	// Find virtual environment modules
	venvs, err := findVirtualEnvironments()
	if err != nil {
		return fmt.Errorf("error finding virtual environments: %w", err)
	}
	// Find modules in each virtual environment
	var modules []PythonModule
	var documents []*common.SearchDocument
	for _, venv := range venvs {
		slog.Info("Found virtual environment", "venv", venv)
		modules, err = findModules(venv)
		if err != nil {
			return fmt.Errorf("error finding modules in virtual environment %s: %w", venv, err)
		}
		// Create a search document for each module
		for _, module := range modules {
			doc := &common.SearchDocument{
				Language: common.Python,
				Name:     module.Name,
				Path:     module.Path,
			}
			documents = append(documents, doc)
		}
	}
	// TODO Find standard library modules
	// Index the modules
	slog.Info("Total", "documents", len(documents))
	err = common.IndexDocuments(db, documents)
	if err != nil {
		return fmt.Errorf("error indexing documents: %w", err)
	}
	return nil
}

func findVirtualEnvironments() ([]string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return nil, fmt.Errorf("home environment variable not set")
	}
	venvs := make([]string, 0)
	filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		if !d.IsDir() {
			name := filepath.Base(path)
			if name == "pyvenv.cfg" {
				venvs = append(venvs, filepath.Dir(path))
				return fs.SkipDir
			}
		}
		return nil
	})
	return venvs, nil
}

func findSitePackagesDir(root string) (string, error) {
	// Find site-packages directory
	var sitePackagesDir string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("Error walking directory", "path", path, "error", err)
			return nil
		}
		if d.IsDir() && d.Name() == "site-packages" {
			sitePackagesDir = path
			return fs.SkipDir
		}
		return nil
	})
	// Check if site-packages directory was found
	if err != nil || sitePackagesDir == "" {
		slog.Warn("Error finding site-packages directory", "error", err)
		return "", err
	}
	// Find libraries in site-packages directory
	return sitePackagesDir, nil
}

func moduleNameFromPath(sitePackagesDir string, path string) string {
	path = strings.ReplaceAll(path, sitePackagesDir, "")
	path = strings.TrimPrefix(path, string(os.PathSeparator))
	path = strings.TrimSuffix(path, string(os.PathSeparator)+"__init__.py")
	path = strings.TrimSuffix(path, ".py")
	path = strings.ReplaceAll(path, string(os.PathSeparator), ".")
	return path
}

type PythonModule struct {
	Venv            string
	Name            string
	Path            string
	SitePackagesDir string
}

func findModules(venv string) ([]PythonModule, error) {
	// Find site-packages directory
	sitePackagesDir, err := findSitePackagesDir(venv)
	if err != nil {
		slog.Error("Error finding sitePackagesDir", "sitePackagesDir", sitePackagesDir)
		return nil, err
	}
	// Find init files (which identify a module)
	initFiles := make([]string, 0)
	err = filepath.Walk(sitePackagesDir,
		func(path string, d os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && d.Name() == "__init__.py" {
				initFiles = append(initFiles, path)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	// Find the other python files in those modules
	allFiles := make([]string, 0)
	for _, initFile := range initFiles {
		dir := filepath.Dir(initFile)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".py") {
				allFiles = append(allFiles, filepath.Join(dir, file.Name()))
			}
		}
	}
	// Map to modules
	acc := make([]PythonModule, 0)
	for _, file := range allFiles {
		name := moduleNameFromPath(sitePackagesDir, file)
		acc = append(acc, PythonModule{
			Venv:            venv,
			Name:            name,
			Path:            file,
			SitePackagesDir: sitePackagesDir,
		})
	}
	return acc, nil
}
