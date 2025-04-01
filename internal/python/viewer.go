package python

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/brandtg/rtfm/internal/common"
)

func readPydoc(module *PythonModule) (string, error) {
	if module == nil {
		return "", fmt.Errorf("module is nil")
	}
	binary := filepath.Join(module.Venv, "bin", "python")
	cmd := exec.Command(binary, "-m", "pydoc", module.Name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running pydoc: %v %s", err, string(output))
	}
	return string(output), nil
}

func View(baseOutputDir string, target string, showSource bool) (*PythonModule, error) {
	outputDir := pythonOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Find the record for pythonClassName
	module, err := fetchPythonModule(db, target)
	if err != nil {
		return nil, err
	}
	if showSource {
		// Output the source code
		source, err := os.ReadFile(module.Path)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(source))
	} else {
		// Output the pydoc
		pydoc, err := readPydoc(module)
		if err != nil {
			return nil, err
		}
		fmt.Println(pydoc)
	}
	return module, nil
}
