package javascript

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/brandtg/rtfm/app/common"
)

func Index() error {
	slog.Info("Indexing Javascript modules...")
	// Connect to the database
	db, err := common.OpenDB()
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	// Find node_modules directories
	nodeModulesDirs, err := findNodeModulesDirs()
	if err != nil {
		slog.Error("Error finding node modules", "error", err)
		return err
	}
	// Find packages in each node_modules directory
	for _, nodeModuleDir := range nodeModulesDirs {
		slog.Info("Found node_modules", "path", nodeModuleDir)
		// Find packages in node_modules
		var packages []*JavaScriptPackage
		packages, err = findJavaScriptPackages(nodeModuleDir)
		if err != nil {
			slog.Error("Error finding JavaScript packages", "error", err)
			return err
		}
		// Find modules in each package
		for _, pkg := range packages {
			// Find files in package
			var files []string
			files, err = findJavaScriptFiles(pkg)
			if err != nil {
				slog.Error("Error finding JavaScript files", "error", err)
				return err
			}
			// Map files to JavaScriptModule
			javascriptModules := make([]JavaScriptModule, 0)
			for _, file := range files {
				javaScriptModule := JavaScriptModule{
					Path:    strings.Replace(file, nodeModuleDir, "", 1)[1:],
					Package: pkg,
				}
				javascriptModules = append(javascriptModules, javaScriptModule)
			}
			// Create a search document for each module
			documents := make([]*common.SearchDocument, 0)
			for _, module := range javascriptModules {
				doc := &common.SearchDocument{
					Language: common.Javascript,
					Name:     module.Path,
					Path:     filepath.Join(pkg.NodeModulesDir, module.Path),
				}
				documents = append(documents, doc)
			}
			// Write to database
			err = common.IndexDocuments(db, documents)
			if err != nil {
				return fmt.Errorf("error indexing documents: %w", err)
			}
		}
	}
	return nil
}

func isNestedNodeModules(path string) bool {
	pathTokens := strings.Split(path, string(os.PathSeparator))
	count := 0
	for _, token := range pathTokens {
		if token == "node_modules" {
			count++
		}
	}
	return count > 1
}

func shouldExclude(path string) bool {
	// TODO Add more exclusions
	return strings.Contains(path, "/anaconda3/")
}

func findNodeModulesDirs() ([]string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return nil, fmt.Errorf("home environment variable not set")
	}
	nodeModules := make([]string, 0)
	err := filepath.WalkDir(home, func(path string, d os.DirEntry, err error) error {
		// Skip hidden directories
		if d.IsDir() && d.Name()[0] == '.' {
			slog.Debug("Skipping hidden directory", "path", path)
			return fs.SkipDir
		}
		// Skip excluded paths
		if shouldExclude(path) {
			slog.Debug("Skipping excluded directory", "path", path)
			return fs.SkipDir
		}
		// Log errors (usually permissions)
		if err != nil {
			slog.Warn("Error walking directory", "path", path, "error", err)
		}
		// Check for node_modules
		if d.IsDir() && d.Name() == "node_modules" && !isNestedNodeModules(path) {
			nodeModules = append(nodeModules, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return nodeModules, nil
}

type JavaScriptPackage struct {
	NodeModulesDir string
	Name           string
	Version        string
	Path           string
}

func (p *JavaScriptPackage) fullPath() string {
	return filepath.Join(p.NodeModulesDir, p.Path)
}

type JavaScriptModule struct {
	Path    string
	Package *JavaScriptPackage
}

func parsePackageJSON(nodeModulesDir string, path string) (*JavaScriptPackage, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	// Parse JSON
	var data map[string]any
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	// Extract metadata
	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract name from JSON")
	}
	version, ok := data["version"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract version from JSON")
	}
	return &JavaScriptPackage{
		NodeModulesDir: nodeModulesDir,
		Name:           name,
		Version:        version,
		Path:           strings.ReplaceAll(path, nodeModulesDir, ""),
	}, nil
}

func findJavaScriptPackages(nodeModulesDir string) ([]*JavaScriptPackage, error) {
	packages := make([]*JavaScriptPackage, 0)
	err := filepath.WalkDir(nodeModulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("Error walking directory", "path", path, "error", err)
			return nil
		}
		// Skip hidden directories
		if d.IsDir() && d.Name()[0] == '.' {
			slog.Debug("Skipping hidden directory", "path", path)
			return fs.SkipDir
		}
		// Check for package.json
		if d.IsDir() {
			packageJSONFile := filepath.Join(path, "package.json")
			if common.Exists(packageJSONFile) {
				pkg, err := parsePackageJSON(nodeModulesDir, packageJSONFile)
				if err != nil {
					slog.Warn("Error parsing package.json", "path", packageJSONFile, "error", err)
					return fs.SkipDir
				}
				packages = append(packages, pkg)
				return fs.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func findJavaScriptFiles(pkg *JavaScriptPackage) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(filepath.Dir(pkg.fullPath()), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Warn("Error walking directory", "path", path, "error", err)
			return nil
		}
		if !d.IsDir() && common.HasAnySuffix(path, ".js", ".ts", ".jsx", ".tsx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
