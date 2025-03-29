package core

import (
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

func Index() {
	outputDir, err := ensureOutputDir()
	if err != nil {
		slog.Error("Error creating output directory", "error", err)
		panic(err)
	}
	findMavenArtifacts(filepath.Join(outputDir, "maven"))
}
