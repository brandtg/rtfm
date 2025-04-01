package common

import (
	"database/sql"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

func MakeFuzzy(pattern string) string {
	re := regexp.MustCompile(`\s+`)
	return "%" + re.ReplaceAllString(pattern, "%") + "%"
}

func OpenDB(dir string, name string) (*sql.DB, error) {
	path := filepath.Join(dir, name)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return db, nil
}
