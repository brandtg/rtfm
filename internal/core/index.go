package core

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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
	slog.Info("Found Maven artifacts", "artifacts", len(mavenArtifacts))
	jdkClasses, err := findJDKClasses()
	if err != nil {
		return fmt.Errorf("error finding JDK classes: %w", err)
	}
	slog.Info("Found JDK classes", "classes", len(jdkClasses))
	for _, jdkClass := range jdkClasses {
		slog.Info("Processing JDK class", "class", jdkClass.path)
	}
	return nil
}

func createDatabase(outputDir string) error {
	dbPath := filepath.Join(outputDir, "example.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS java_classes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	slog.Info("Database and table created successfully", "path", dbPath)
	return nil
}

func Index() {
	outputDir, err := ensureOutputDir()
	if err != nil {
		slog.Error("Error creating output directory", "error", err)
		panic(err)
	}
	err = createDatabase(outputDir)
	if err != nil {
		slog.Error("Error creating database", "error", err)
		panic(err)
	}
	indexJava(outputDir)
}
