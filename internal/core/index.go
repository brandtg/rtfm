package core

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
    "fmt"
    "strings"
)

type MavenCoordinates struct {
    GroupId string
    ArtifactId string
    Version string
    Classifier string
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
				GroupId:    strings.Join(groupParts, "."),
				ArtifactId: artifactId,
				Version:    version,
				Classifier: classifier,
			}, nil
        }
    }
    return nil, fmt.Errorf("invalid Maven path: %s", path)
}

func visitMaven(path string, d fs.DirEntry, err error) error {
    if err != nil {
        return err
    }
    if !d.IsDir() {
        if strings.HasSuffix(path, "-javadoc.jar") || strings.HasSuffix(path, "-sources.jar") {
            c, err := parseMavenCoordinates(path)
            if err != nil {
                return err
            }
            slog.Info("Parsed coordinates", "coordinates", c)
        }
    }
    return nil
}

func Index() {
    home, err := os.UserHomeDir()
    if err != nil {
        panic(err)
    }
    mavenDir := filepath.Join(home, ".m2")
    slog.Info("Maven directory", "mavenDir", mavenDir)
    err = filepath.WalkDir(mavenDir, visitMaven)
    if err != nil {
        panic(err)
    }
}
