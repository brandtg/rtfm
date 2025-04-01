package python

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/brandtg/rtfm/internal/common"
)

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS python_module (
        key TEXT PRIMARY KEY,
        venv TEXT,
        name TEXT,
        path TEXT,
        site_packages_dir TEXT
    )
    `)
	if err != nil {
		return err
	}
	return nil
}

func insertPythonModules(db *sql.DB, modules []PythonModule) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO python_module (
            key,
            venv,
            name,
            path,
            site_packages_dir
        )
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, module := range modules {
		_, err := stmt.Exec(
			module.key(),
			module.Venv,
			module.Name,
			module.Path,
			module.SitePackagesDir,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func listPythonModules(db *sql.DB, venv string, pattern string, exact bool) ([]PythonModule, error) {
	if !exact {
		pattern = common.MakeFuzzy(pattern)
	}
	rows, err := db.Query(`
        SELECT 
            venv
            , name
            , path
            , site_packages_dir
        FROM 
            python_module
        WHERE 
            LOWER(name) LIKE ?
            AND (? OR venv = ?)
    `, pattern, venv == "", venv)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var modules []PythonModule
	for rows.Next() {
		var module PythonModule
		err := rows.Scan(&module.Venv, &module.Name, &module.Path, &module.SitePackagesDir)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func fetchPythonModule(db *sql.DB, target string) (*PythonModule, error) {
	tokens := strings.Split(target, ":")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid target format: %s", target)
	}
	rows, err := db.Query(`
        SELECT venv, name, path, site_packages_dir
        FROM python_module
        WHERE venv = ? AND name = ?
    `, tokens[0], tokens[1])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var modules []PythonModule
	for rows.Next() {
		var module PythonModule
		err := rows.Scan(&module.Venv, &module.Name, &module.Path, &module.SitePackagesDir)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	if len(modules) > 0 {
		return &modules[0], nil
	}
	return nil, fmt.Errorf("module not found: %s", target)
}
