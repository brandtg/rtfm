package common

import (
	"os"
	"path/filepath"
)

func EnsureOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	outputDir := filepath.Join(home, ".local", "share", "rtfm")
	err = os.MkdirAll(outputDir, 0o755)
	if err != nil {
		panic(err)
	}
	return outputDir
}
