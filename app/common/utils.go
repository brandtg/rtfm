package common

import (
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func Exists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func HasAnySuffix(filename string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}
