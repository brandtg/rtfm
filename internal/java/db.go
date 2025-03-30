package java

import (
	"database/sql"
	"log/slog"
	"path/filepath"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(outputDir string) (*sql.DB, error) {
	path := filepath.Join(outputDir, "java.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS java_class (
        name TEXT PRIMARY KEY,
        path TEXT,
        source TEXT,
        artifact_path TEXT,
        artifact_group_id TEXT,
        artifact_id TEXT,
        artifact_version TEXT
    )
    `)
	if err != nil {
		return err
	}
	return nil
}

func InsertJavaClasses(db *sql.DB, docs []JavaClass) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO java_class (
            name,
            path,
            source,
            artifact_path,
            artifact_group_id,
            artifact_id,
            artifact_version
        )
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, doc := range docs {
		_, err := stmt.Exec(
			doc.Name,
			doc.Path,
			doc.Source,
			doc.Artifact.Path,
			doc.Artifact.GroupId,
			doc.Artifact.ArtifactId,
			doc.Artifact.Version)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func makeFuzzy(pattern string) string {
	re := regexp.MustCompile(`\s+`)
	return "%" + re.ReplaceAllString(pattern, "%") + "%"
}

func ListJavaClasses(
	db *sql.DB,
	pattern string,
	group string,
	artifact string,
	version string,
	exact bool,
) ([]JavaClass, error) {
	if !exact {
		pattern = makeFuzzy(pattern)
	}
	slog.Debug("Searching for Java classes", "pattern", pattern)
	rows, err := db.Query(`
        SELECT 
            name
            , path
            , source
            , artifact_path
            , artifact_group_id
            , artifact_id
            , artifact_version 
        FROM 
            java_class 
        WHERE 
            LOWER(name) LIKE ?
            AND (? OR artifact_group_id = ?)
            AND (? OR artifact_id = ?)
            AND (? OR artifact_version = ?)
        ORDER BY
            name
        `, pattern, group == "", group, artifact == "", artifact, version == "", version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var javaClasses []JavaClass
	for rows.Next() {
		var doc JavaClass
		doc.Artifact = &MavenCoordinates{}
		err := rows.Scan(
			&doc.Name,
			&doc.Path,
			&doc.Source,
			&doc.Artifact.Path,
			&doc.Artifact.GroupId,
			&doc.Artifact.ArtifactId,
			&doc.Artifact.Version)
		if err != nil {
			return nil, err
		}
		javaClasses = append(javaClasses, doc)
	}
	return javaClasses, nil
}
