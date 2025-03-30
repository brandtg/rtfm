package java

import (
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func View(baseOutputDir string, target string, serverPort int) (*JavaClass, error) {
	outputDir := javaOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := openDB(outputDir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Find the record for javaClassName
	javaClass, err := fetchJavaClass(db, target)
	if err != nil {
		return nil, err
	}
	// Format the Javadoc as Markdown
	markdown := FormatMarkdown(filepath.Join(outputDir, javaClass.Path))
	fmt.Println(markdown)
	return javaClass, nil
}
