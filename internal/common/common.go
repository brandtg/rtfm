package common

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func Dedupe(input []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func Exists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func AtLeastOneFileExists(files []string) bool {
	return slices.ContainsFunc(files, Exists)
}

func HasAnySuffix(filename string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}
