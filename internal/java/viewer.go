package java

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/brandtg/rtfm/internal/common"
)

func readJavaSource(outputDir string, javaClass *JavaClass) (string, error) {
	if strings.HasPrefix(javaClass.Source, "http") {
		// Check if we've cached the file
		cachedFile := filepath.Join(outputDir, "sources", "_jdk", javaClass.Name)
		if common.Exists(cachedFile) {
			// Read it
			data, err := os.ReadFile(cachedFile)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		// Fetch the HTML from the web
		resp, err := http.Get(javaClass.Source)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		// Cache the result
		err = os.MkdirAll(filepath.Dir(cachedFile), 0755)
		if err != nil {
			slog.Warn("Failed to create javadoc cache directory", "path", cachedFile, "error", err)
		}
		err = os.WriteFile(cachedFile, bodyBytes, 0644)
		if err != nil {
			slog.Warn("Failed to cache javadoc HTML", "path", javaClass.Path, "error", err)
		}
		return string(bodyBytes), nil
	}
	// Read HTML from the file
	source, err := os.ReadFile(filepath.Join(outputDir, javaClass.Source))
	if err != nil {
		return "", err
	}
	return string(source), nil
}

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
		source, err := readJavaSource(outputDir, javaClass)
		if err != nil {
			return nil, err
		}
		code, err := common.HighlightCode(string(source), "java")
		if err != nil {
			return nil, err
		}
		fmt.Println(code)
	} else {
		// Format the Javadoc as Markdown
		markdown, err := FormatMarkdown(outputDir, javaClass)
		if err != nil {
			return nil, err
		}
		fmt.Println(markdown)
	}
	return javaClass, nil
}
