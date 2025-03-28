package core

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
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

func (m *MavenCoordinates) Dir() string {
	return filepath.Join(
		m.Classifier,
		strings.ReplaceAll(m.GroupId, ".", string(filepath.Separator)),
		m.ArtifactId, m.Version)
}

func parseMavenCoordinates(path string) (*MavenCoordinates, error) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "repository" && i+3 < len(parts) {
			groupParts := parts[i+1 : len(parts)-3]
			artifactId := parts[len(parts)-3]
			version := parts[len(parts)-2]
			filename := filepath.Base(path)
			filename = strings.TrimSuffix(filename, filepath.Ext(filename)) // drop .jar/.pom
			prefix := artifactId + "-" + version
			classifier := ""
			if strings.HasPrefix(filename, prefix+"-") {
				classifier = strings.TrimPrefix(filename, prefix+"-")
			}
			return &MavenCoordinates{
				Path:       path,
				GroupId:    strings.Join(groupParts, "."),
				ArtifactId: artifactId,
				Version:    version,
				Classifier: classifier,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid Maven path: %s", path)
}

func extractMavenArtifact(coordinates *MavenCoordinates, outputDir string) error {
	extractDir := filepath.Join(outputDir, coordinates.Dir())
	if _, err := os.Stat(extractDir); err == nil {
		slog.Debug("Directory already exists, skipping extraction", "dir", extractDir)
		return nil
	}
	err := os.MkdirAll(extractDir, 0o755)
	if err != nil {
		return err
	}
	r, err := zip.OpenReader(coordinates.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(extractDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func findMavenArtifacts(mavenDir string, outputDir string) error {
	err := filepath.WalkDir(mavenDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.HasSuffix(path, "-javadoc.jar") || strings.HasSuffix(path, "-sources.jar") {
				c, err := parseMavenCoordinates(path)
				if err != nil {
					return err
				}
				err = extractMavenArtifact(c, outputDir)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}
