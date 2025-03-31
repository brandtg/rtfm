package java

import (
	"path/filepath"
	"strings"
)

type Link struct {
	Text  string
	Title string
	Href  string
}

type MavenCoordinates struct {
	Path       string
	GroupId    string
	ArtifactId string
	Version    string
	Classifier string
}

func (m *MavenCoordinates) OutputDir() string {
	return filepath.Join(
		m.Classifier,
		strings.ReplaceAll(m.GroupId, ".", string(filepath.Separator)),
		m.ArtifactId, m.Version)
}

type JavaClass struct {
	Name     string
	Path     string
	Source   string
	Artifact *MavenCoordinates
}

func javaOutputDir(baseOutputDir string) string {
	return filepath.Join(baseOutputDir, "java")
}
