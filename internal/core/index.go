package core

import (
    "fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func ensureOutputDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	outputDir := filepath.Join(home, ".rtfm")
	err = os.MkdirAll(outputDir, 0o755)
	if err != nil {
		return "", err
	}
	return outputDir, nil
}

func indexJava(outputDir string) error {
    mavenArtifacts, err := findMavenArtifacts(filepath.Join(outputDir, "maven"))
    if err != nil {
        return fmt.Errorf("error finding Maven artifacts: %w", err)
    }
    _ = mavenArtifacts
    // slog.Info("Found Maven artifacts", "artifacts", mavenArtifacts)
    jdkClasses, err := findJDKClasses()
    if err != nil {
        return fmt.Errorf("error finding JDK classes: %w", err)
    }
    _ = jdkClasses
    // slog.Info("Found JDK classes", "classes", jdkClasses)
    for _, jdkClass := range jdkClasses {
        slog.Info("Processing JDK class", "class", jdkClass.path)
    }
    return nil
}

func Index() {
	outputDir, err := ensureOutputDir()
	if err != nil {
		slog.Error("Error creating output directory", "error", err)
		panic(err)
	}
	indexJava(outputDir)
}
