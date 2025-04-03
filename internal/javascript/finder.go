package javascript

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/brandtg/rtfm/internal/common"
)

func Find(
	baseOutputDir string,
	pattern string,
	nodeModulesDir string,
	format string,
	exact bool,
) ([]JavaScriptModule, error) {
	outputDir := javascriptOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Resolve nodeModulesDir if it's not absolute
	if nodeModulesDir != "" {
		nodeModulesDir, err = filepath.Abs(nodeModulesDir)
		if err != nil {
			slog.Error("Error resolving nodeModulesDir", "error", err)
			return nil, err
		}
	}
	// Search for JavaScript modules matching pattern
	javascriptModules, err := listJavascriptModules(db, nodeModulesDir, pattern, exact)
	if err != nil {
		return nil, err
	}
	// Print the results
	switch format {
	case "default":
		formatted := make([]string, len(javascriptModules))
		for i, javascriptModule := range javascriptModules {
			formatted[i] = javascriptModule.key()
		}
		sort.Strings(formatted)
		for _, javascriptModule := range formatted {
			fmt.Println(javascriptModule)
		}
	case "json":
		json, err := json.MarshalIndent(javascriptModules, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(json))
	case "source":
		for _, javascriptModule := range javascriptModules {
			fmt.Println(javascriptModule.Path)
		}
	default:
		slog.Error("Unsupported format:", "format", format)
	}
	return javascriptModules, nil
}
