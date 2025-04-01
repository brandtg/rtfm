package java

import (
	"fmt"
	"path/filepath"
	"strings"
)

const DB_NAME = "java.db"

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

func (j *JavaClass) key() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		j.Artifact.GroupId,
		j.Artifact.ArtifactId,
		j.Artifact.Version,
		j.Name)
}
