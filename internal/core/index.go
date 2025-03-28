package core

import (
	"log/slog"
	"os"
	"path/filepath"
)

func Index() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	mavenDir := filepath.Join(home, ".m2")
	slog.Info("Maven directory", "mavenDir", mavenDir)
	outputDir := filepath.Join(home, "rtfm")
	os.MkdirAll(outputDir, 0o755)
    slog.Info("Output directory", "outputDir", outputDir)
    err = findMavenArtifacts(mavenDir, outputDir)
    if err != nil {
        slog.Error("Error finding maven artifacts", "error", err)
        panic(err)
    }
}
