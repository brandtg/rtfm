package javascript

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brandtg/rtfm/internal/common"
)

func View(baseOutputDir string, target string, showSource bool) (*JavaScriptModule, error) {
	outputDir := javascriptOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Find the record for target
	javaScriptModule, err := fetchJavaScriptModule(db, target)
	if err != nil {
		return nil, err
	}
	if showSource {
		// Output the source code
		source, err := os.ReadFile(filepath.Join(javaScriptModule.Package.NodeModulesDir, javaScriptModule.Path))
		if err != nil {
			return nil, err
		}
		fmt.Println(string(source))
	} else {
		// Output the documentation (TODO - needs to be computed first)
		fmt.Println("Documentation not yet implemented")
	}
	return nil, nil
}
