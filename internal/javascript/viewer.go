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
	// Compute full path
	path := filepath.Join(javaScriptModule.Package.NodeModulesDir, javaScriptModule.Path)
	if showSource {
		// Output the source code
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		code, err := common.HighlightCode(string(source), "javascript")
		if err != nil {
			return nil, err
		}
		fmt.Println(code)
	} else {
		// Output the JSDoc
		doc, err := ParseJSDoc(path)
		if err != nil {
			return nil, err
		}
		fmt.Println(doc)
	}
	return nil, nil
}
