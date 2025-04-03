package python

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
	venv string,
	format string,
	exact bool,
) ([]PythonModule, error) {
	outputDir := pythonOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Resolve venv if it's not absolute
	if venv != "" {
		venv, err = filepath.Abs(venv)
		if err != nil {
			slog.Error("Error resolving venv", "error", err)
			return nil, err
		}
	}
	// Search for Python modules matching pattern
	pythonModules, err := listPythonModules(db, venv, pattern, exact)
	if err != nil {
		return nil, err
	}
	// Print the results
	switch format {
	case "default":
		formatted := make([]string, len(pythonModules))
		for i, pythonModule := range pythonModules {
			formatted[i] = pythonModule.key()
		}
		sort.Strings(formatted)
		for _, pythonModule := range formatted {
			fmt.Println(pythonModule)
		}
	case "json":
		json, err := json.MarshalIndent(pythonModules, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(json))
	case "source":
		for _, pythonModule := range pythonModules {
			fmt.Println(pythonModule.Path)
		}
	case "module":
		for _, pythonModule := range pythonModules {
			fmt.Println(pythonModule.Name)
		}
	default:
		slog.Error("Unsupported format:", "format", format)
	}
	return pythonModules, nil
}
