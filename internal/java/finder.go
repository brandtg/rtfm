package java

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
)

func Find(
	baseOutputDir string,
	pattern string,
	group string,
	artifact string,
	version string,
	exact bool,
	format string,
) ([]JavaClass, error) {
	outputDir := javaOutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := openDB(outputDir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Search for Java classes matching pattern
	javaClasses, err := listJavaClasses(db, pattern, group, artifact, version, exact)
	if err != nil {
		return nil, err
	}
	// Print the results
	switch format {
	case "default":
		formatted := make([]string, len(javaClasses))
		for i, javaClass := range javaClasses {
			formatted[i] = fmt.Sprintf(
				"%s:%s:%s:%s",
				javaClass.Artifact.GroupId,
				javaClass.Artifact.ArtifactId,
				javaClass.Artifact.Version,
				javaClass.Name)
		}
		sort.Strings(formatted)
		for _, javaClass := range formatted {
			fmt.Println(javaClass)
		}
	case "json":
		json, err := json.MarshalIndent(javaClasses, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(json))
	case "javadoc":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Path)
		}
	case "source":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Source)
		}
	case "class":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Name)
		}
	default:
		slog.Error("Unknown format", "format", format)
	}
	return javaClasses, nil
}
