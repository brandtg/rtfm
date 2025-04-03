package javascript

import (
	"fmt"
	"path/filepath"
)

const DB_NAME = "javascript.db"

type JavaScriptPackage struct {
	NodeModulesDir string
	Name           string
	Version        string
	Path           string
}

func (p *JavaScriptPackage) fullPath() string {
	return filepath.Join(p.NodeModulesDir, p.Path)
}

type JavaScriptModule struct {
	Path    string
	Package *JavaScriptPackage
}

func (m *JavaScriptModule) key() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		m.Package.NodeModulesDir,
		m.Package.Name,
		m.Package.Version,
		m.Path)
}
