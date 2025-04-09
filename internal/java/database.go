package java

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/brandtg/rtfm/internal/common"
)

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS java_class (
        key TEXT PRIMARY KEY,
        name TEXT,
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

func insertJavaClasses(db *sql.DB, docs []JavaClass) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO java_class (
            key,
            name,
            path,
            source,
            artifact_path,
            artifact_group_id,
            artifact_id,
            artifact_version
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, doc := range docs {
		_, err := stmt.Exec(
			doc.key(),
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

func listJavaClasses(
	db *sql.DB,
	pattern string,
	group string,
	artifact string,
	version string,
	exact bool,
) ([]JavaClass, error) {
	if !exact {
		pattern = common.MakeFuzzy(pattern)
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

func fetchJavaClass(db *sql.DB, target string) (*JavaClass, error) {
	// tokens := strings.Split(target, ":")
	// if len(tokens) != 4 {
	// 	return nil, fmt.Errorf("invalid target format: %s", target)
	// }
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
			key = ?
    `, target)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var docs []JavaClass
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
		docs = append(docs, doc)
	}
	if len(docs) > 0 {
		return &docs[0], nil
	}
	return nil, fmt.Errorf("java class not found: %s", target)
}
