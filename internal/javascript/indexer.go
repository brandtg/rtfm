package javascript

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/brandtg/rtfm/internal/common"
)

func javascriptOutputDir(baseOutputDir string) string {
	dir := filepath.Join(baseOutputDir, "javascript")
	os.MkdirAll(dir, os.ModePerm)
	return dir
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

func Index(baseOutputDir string) error {
	// Connect to SQLite database
	outputDir := javascriptOutputDir(baseOutputDir)
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return err
	}
	defer db.Close()
	err = createTables(db)
	if err != nil {
		return err
	}
	// Find node_modules directories
	nodeModulesDirs, err := findNodeModulesDirs()
	if err != nil {
		slog.Error("Error finding node modules", "error", err)
		return err
	}
	for _, nodeModuleDir := range nodeModulesDirs {
		slog.Info("Found node_modules directory", "path", nodeModuleDir)
		// Find packages in node_modules
		packages, err := findJavaScriptPackages(nodeModuleDir)
		if err != nil {
			slog.Error("Error finding JavaScript packages", "error", err)
			return err
		}
		for _, pkg := range packages {
			// Find files in package
			files, err := findJavaScriptFiles(pkg)
			if err != nil {
				slog.Error("Error finding JavaScript files", "error", err)
				return err
			}
			slog.Debug("Found count of JavaScript files", "count", len(files))
			// Map files to JavaScriptModule
			javascriptModules := make([]JavaScriptModule, 0)
			for _, file := range files {
				javaScriptModule := JavaScriptModule{
					Path:    strings.Replace(file, nodeModuleDir, "", 1)[1:],
					Package: pkg,
				}
				javascriptModules = append(javascriptModules, javaScriptModule)
			}
			// Write to database
			err = insertJavascriptModules(db, javascriptModules)
			if err != nil {
				slog.Error("Error inserting JavaScript modules", "error", err)
				return err
			}
		}
	}
	return nil
}
