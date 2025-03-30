package java

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

func parseClassifier(path string, artifactId string, version string) string {
	prefix := artifactId + "-" + version
	filename := filepath.Base(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) // drop .jar/.pom
	if strings.HasPrefix(filename, prefix+"-") {
		return strings.TrimPrefix(filename, prefix+"-")
	}
	return ""
}

func parsePath(path string) (*MavenCoordinates, error) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "repository" && i+3 < len(parts) {
			groupId := strings.Join(parts[i+1:len(parts)-3], ".")
			artifactId := parts[len(parts)-3]
			version := parts[len(parts)-2]
			classifier := parseClassifier(path, artifactId, version)
			return &MavenCoordinates{
				Path:       path,
				GroupId:    groupId,
				ArtifactId: artifactId,
				Version:    version,
				Classifier: classifier,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid Maven path: %s", path)
}

func ExtractArtifact(coords *MavenCoordinates, outputDir string) (bool, error) {
	// Create output directory
	dest := filepath.Join(outputDir, coords.OutputDir())
	if _, err := os.Stat(dest); err == nil {
		return false, nil
	}
	err := os.MkdirAll(dest, 0o755)
	if err != nil {
		return false, err
	}
	// Open the archive
	r, err := zip.OpenReader(coords.Path)
	if err != nil {
		return false, err
	}
	defer r.Close()
	// Extract files
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		// Create archive directory
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return false, err
			}
			continue
		}
		// Create file's parent directory
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return false, err
		}
		// Create file
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return false, err
		}
		// Copy file content
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return false, err
		}
		_, err = io.Copy(outFile, rc)
		// Cleanup
		outFile.Close()
		rc.Close()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func shouldExtract(path string) bool {
	return strings.HasSuffix(path, "-javadoc.jar") ||
		strings.HasSuffix(path, "-sources.jar")
}

func DiscoverArtifacts(inputDir string) ([]*MavenCoordinates, error) {
	acc := []*MavenCoordinates{}
	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && shouldExtract(path) {
			coords, err := parsePath(path)
			if err != nil {
				return err
			}
			acc = append(acc, coords)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func ListRepos() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	// TODO Add gradle
	mavenDir := filepath.Join(home, ".m2", "repository")
	return []string{mavenDir}, nil
}
