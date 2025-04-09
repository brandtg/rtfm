package java

import (
	"fmt"
	"net/url"
	"os"
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
	dir := filepath.Join(baseOutputDir, "java")
	os.MkdirAll(dir, os.ModePerm)
	return dir
}

func (j *JavaClass) key() string {
	// JDK classes
	if strings.HasPrefix(j.Path, "https://docs.oracle.com") {
		url, err := url.Parse(j.Path)
		if err != nil {
			panic(err)
		}
		version := strings.Split(strings.Replace(url.Path, "/en/java/javase/", "", 1), "/")[0]
		return fmt.Sprintf(
			"jdk:java.base:%s:%s",
			version,
			j.Name)
	}
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		j.Artifact.GroupId,
		j.Artifact.ArtifactId,
		j.Artifact.Version,
		j.Name)
}
