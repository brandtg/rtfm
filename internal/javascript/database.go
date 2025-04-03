package javascript

import (
	"database/sql"
	"fmt"

	"github.com/brandtg/rtfm/internal/common"
)

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS javascript_module (
        key TEXT PRIMARY KEY,
		node_modules TEXT,
		package_name TEXT,
		package_version TEXT,
		path TEXT
    )
    `)
	if err != nil {
		return err
	}
	return nil
}

func insertJavascriptModules(db *sql.DB, modules []JavaScriptModule) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO javascript_module (
            key,
			node_modules,
			package_name,
			package_version,
            path
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
			module.Package.NodeModulesDir,
			module.Package.Name,
			module.Package.Version,
			module.Path,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func listJavascriptModules(
	db *sql.DB,
	nodeModulesDir string,
	pattern string,
	exact bool,
) ([]JavaScriptModule, error) {
	if !exact {
		pattern = common.MakeFuzzy(pattern)
	}
	rows, err := db.Query(`
        SELECT
			node_modules,
			package_name,
			package_version,
            path
        FROM
            javascript_module
        WHERE
            LOWER(path) LIKE ?
            AND (? OR node_modules = ?)
    `, pattern, nodeModulesDir == "", nodeModulesDir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var modules []JavaScriptModule
	for rows.Next() {
		var module JavaScriptModule
		module.Package = &JavaScriptPackage{}
		err := rows.Scan(
			&module.Package.NodeModulesDir,
			&module.Package.Name,
			&module.Package.Version,
			&module.Path,
		)
		if err != nil {
			return nil, err
		}
		modules = append(modules, module)
	}
	return modules, nil
}

func fetchJavaScriptModule(
	db *sql.DB,
	target string,
) (*JavaScriptModule, error) {
	rows, err := db.Query(`
        SELECT
			node_modules
			, package_name
			, package_version
            , path
        FROM
            javascript_module
        WHERE
			key = ?
	`, target)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var acc []JavaScriptModule
	for rows.Next() {
		var module JavaScriptModule
		module.Package = &JavaScriptPackage{}
		err := rows.Scan(
			&module.Package.NodeModulesDir,
			&module.Package.Name,
			&module.Package.Version,
			&module.Path,
		)
		if err != nil {
			return nil, err
		}
		acc = append(acc, module)
	}
	if len(acc) == 0 {
		return nil, fmt.Errorf("module not found: %s", target)
	}
	return &acc[0], nil
}
