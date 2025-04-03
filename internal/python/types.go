package python

import "fmt"

const DB_NAME = "python.db"

type PythonModule struct {
	Venv            string
	Name            string
	Path            string
	SitePackagesDir string
}

func (m *PythonModule) key() string {
	return fmt.Sprintf("%s:%s", m.Venv, m.Name)
}
