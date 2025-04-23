package common

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func getOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(home, ".local", "share", "rtfm")
}

func RemoveOutputDir() error {
	outputDir := getOutputDir()
	slog.Info("Removing output directory", "dir", outputDir)
	err := os.RemoveAll(outputDir)
	if err != nil {
		return err
	}
	return nil
}

func EnsureOutputDir() (string, error) {
	outputDir := getOutputDir()
	err := os.MkdirAll(outputDir, 0o755)
	if err != nil {
		return "", err
	}
	return outputDir, nil
}

func OpenDB() (*sql.DB, error) {
	// Create the application data directory if it doesn't exist
	dir, err := EnsureOutputDir()
	if err != nil {
		return nil, err
	}
	// Create the database file if it doesn't exist
	path := filepath.Join(dir, "rtfm.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS code (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			language INTEGER,
			name TEXT,
			path TEXT,
			UNIQUE(language, name, path) ON CONFLICT IGNORE
		)
	`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func IndexDocuments(db *sql.DB, documents []*SearchDocument) error {
	// Create a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Prepare the statement
	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO code (language, name, path)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	// Insert the documents
	for _, doc := range documents {
		_, err := stmt.Exec(doc.Language, doc.Name, doc.Path)
		if err != nil {
			return fmt.Errorf("failed to insert document: %w", err)
		}
	}
	return tx.Commit()
}

func FindDocuments(db *sql.DB, language Language, query string, exact bool) ([]*SearchDocument, error) {
	// Prepare the statement
	stmt, err := db.Prepare(`
		SELECT language, name, path
		FROM code
		WHERE (? = -1 OR language = ?)
		  AND name LIKE ?
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	// Execute the statement
	if !exact {
		query = MakeFuzzy(query)
	}
	rows, err := stmt.Query(language, language, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	// Map the results to SearchDocument
	var documents []*SearchDocument
	for rows.Next() {
		var doc SearchDocument
		err := rows.Scan(&doc.Language, &doc.Name, &doc.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		documents = append(documents, &doc)
	}
	return documents, nil
}
