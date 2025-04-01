package java

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brandtg/rtfm/internal/common"
)

func View(baseOutputDir string, target string, showSource bool) (*JavaClass, error) {
	outputDir := javaOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Find the record for javaClassName
	javaClass, err := fetchJavaClass(db, target)
	if err != nil {
		return nil, err
	}
	if showSource {
		// Output the source code
		source, err := os.ReadFile(filepath.Join(outputDir, javaClass.Source))
		if err != nil {
			return nil, err
		}
		fmt.Println(string(source))
	} else {
		// Format the Javadoc as Markdown
		markdown := FormatMarkdown(outputDir, javaClass)
		fmt.Println(markdown)
	}
	return javaClass, nil
}
