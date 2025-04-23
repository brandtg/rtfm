// Copyright 2025 Greg Brandt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package java

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brandtg/rtfm/app/common"
)

func Index() error {
	slog.Info("Indexing Java classes...")
	// Connect to the database
	db, err := common.OpenDB()
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	// Extract Java classes from the artifacts
	slog.Info("Processing Java artifacts...")
	err = processJavaArtifacts()
	if err != nil {
		return fmt.Errorf("error processing Java artifacts: %w", err)
	}
	// JDK Classes
	slog.Info("Processing JDK classes...")
	err = processJDKClasses()
	if err != nil {
		return fmt.Errorf("error processing JDK classes: %w", err)
	}
	// Index the class files
	slog.Info("Indexing class files...")
	err = indexClassFiles(db)
	if err != nil {
		return fmt.Errorf("error indexing class files: %w", err)
	}
	slog.Info("Java indexing complete")
	return nil
}

func listRepos() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	// TODO Add gradle
	mavenDir := filepath.Join(home, ".m2", "repository")
	return []string{mavenDir}, nil
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

func shouldExtract(path string) bool {
	return strings.HasSuffix(path, "-javadoc.jar") ||
		strings.HasSuffix(path, "-sources.jar")
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

func discoverArtifacts(inputDir string) ([]*MavenCoordinates, error) {
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

func extractZipFile(src string, dst string) error {
	// Open the archive
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	// Extract files
	for _, f := range r.File {
		path := filepath.Join(dst, f.Name)
		// Create archive directory
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			continue
		}
		// Create file's parent directory
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
		// Create file
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		// Copy file content
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		// Cleanup
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractArtifact(coords *MavenCoordinates, outputDir string) error {
	slog.Debug("Extracting artifact", "artifact", coords.Path)
	// Create output directory
	dest := filepath.Join(outputDir, coords.OutputDir())
	if _, err := os.Stat(dest); err == nil {
		return nil
	}
	err := os.MkdirAll(dest, 0o755)
	if err != nil {
		return err
	}
	// Extract the artifact
	err = extractZipFile(coords.Path, dest)
	if err != nil {
		return err
	}
	return nil
}

func javaOutputDir() (string, error) {
	baseOutputDir, err := common.EnsureOutputDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(baseOutputDir, "java")
	os.MkdirAll(dir, os.ModePerm)
	return dir, nil
}

func processJavaArtifacts() error {
	repos, err := listRepos()
	if err != nil {
		return fmt.Errorf("error listing repositories: %w", err)
	}
	outputDir, err := javaOutputDir()
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}
	for _, repo := range repos {
		slog.Debug("Indexing repo", "repo", repo)
		artifacts, err := discoverArtifacts(repo)
		if err != nil {
			slog.Error("Error discovering artifacts", "repo", repo, "error", err)
			return err
		}
		for _, artifact := range artifacts {
			err = extractArtifact(artifact, outputDir)
			if err != nil {
				slog.Error("Error extracting artifact", "artifact", artifact, "error", err)
				return err
			}
		}
	}
	return nil
}

func findJavaVersion() (string, error) {
	cmd := exec.Command("java", "-version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`version "(\d+)\.`)
	match := re.FindStringSubmatch(string(out))
	if len(match) < 2 {
		return "", fmt.Errorf("could not find Java version in output: %s", string(out))
	}
	version := match[1]
	slog.Debug("Found Java version", "version", version)
	return version, nil
}

func findJDKSourceArchive(version string) (string, error) {
	root := "/usr/lib/jvm"
	files, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	prefix := "java-" + version
	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
			// Check if the src.zip file exists
			srcZipPath := filepath.Join(root, file.Name(), "lib", "src.zip")
			if _, err := os.Stat(srcZipPath); err == nil {
				slog.Debug("Found JDK source archive", "archive", srcZipPath)
				return srcZipPath, nil
			}
		}
	}
	return "", fmt.Errorf("JDK source archive not found for version %s", version)
}

func extractJDKSourceArchive(archive string) error {
	outputDir, err := javaOutputDir()
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}
	jdkOutputDir := filepath.Join(outputDir, "sources", "jdk")
	err = os.MkdirAll(jdkOutputDir, 0o755)
	if err != nil {
		return fmt.Errorf("error creating JDK output directory: %w", err)
	}
	err = extractZipFile(archive, jdkOutputDir)
	if err != nil {
		return fmt.Errorf("error extracting JDK source archive: %w", err)
	}
	slog.Debug("Extracted JDK source archive", "archive", archive, "outputDir", jdkOutputDir)
	return nil
}

func processJDKClasses() error {
	// Find the JDK version
	version, err := findJavaVersion()
	if err != nil {
		return fmt.Errorf("error finding Java version: %w", err)
	}
	// Find the JDK source archive
	archive, err := findJDKSourceArchive(version)
	if err != nil {
		return fmt.Errorf("error finding JDK source archive: %w", err)
	}
	// Extract the JDK source archive
	err = extractJDKSourceArchive(archive)
	if err != nil {
		return fmt.Errorf("error extracting JDK source archive: %w", err)
	}
	return nil
}

var (
	packageNameRegex = regexp.MustCompile(`(?m)package\s+([a-zA-Z0-9_.]+);`)
	classNameRegex   = regexp.MustCompile(`(?m)(?:class|interface|record|enum)\s+([a-zA-Z0-9_$]+)`)
	jdkPathPart      = string(filepath.Separator) + "jdk" + string(filepath.Separator)
)

func parseJdkPackageName(path string) (string, error) {
	dir, _ := filepath.Split(path)
	parts := strings.Split(dir, string(filepath.Separator))
	acc := make([]string, 0)
	collect := false
	i := 0
	for i < len(parts) {
		if parts[i] == "jdk" {
			i += 2
			collect = true
			continue
		}
		if collect && len(parts[i]) > 0 {
			acc = append(acc, parts[i])
		}
		i++
	}
	return strings.Join(acc, "."), nil
}

func parseJavaPackageName(path, code string) (string, error) {
	if strings.Contains(path, jdkPathPart) {
		return parseJdkPackageName(path)
	}
	packageNameMatch := packageNameRegex.FindStringSubmatch(code)
	if len(packageNameMatch) < 2 {
		return "", fmt.Errorf("no package name found in %s", path)
	}
	return packageNameMatch[1], nil
}

func parseJavaClassNames(path string) ([]string, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	code := string(data)
	// Parse the package
	packageName, err := parseJavaPackageName(path, code)
	if err != nil {
		return nil, err
	}
	// Parse all class names
	classNames := make([]string, 0)
	classNameMatches := classNameRegex.FindAllStringSubmatch(string(data), -1)
	for _, match := range classNameMatches {
		if len(match) < 2 {
			continue
		}
		className := match[1]
		fullClassName := fmt.Sprintf("%s.%s", packageName, className)
		classNames = append(classNames, fullClassName)
	}
	return classNames, nil
}

func indexClassFiles(db *sql.DB) error {
	// Create the application data directory if it doesn't exist
	outputDir, err := javaOutputDir()
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}
	// Find all Java class files
	sourceDir := filepath.Join(outputDir, "sources")
	documents := make([]*common.SearchDocument, 0)
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		// Ignore non-code files
		fileName := filepath.Base(path)
		if fileName == "package-info.java" || fileName == "module-info.java" {
			return nil
		}
		// Process Java files
		if strings.HasSuffix(path, ".java") {
			names, err := parseJavaClassNames(path)
			if err != nil {
				slog.Error("Error parsing Java class name", "path", path, "error", err)
				return nil
			}
			if len(names) == 0 {
				slog.Info("No class names found", "path", path)
				return nil
			}
			for _, name := range names {
				document := &common.SearchDocument{
					Language: common.Java,
					Name:     name,
					Path:     path,
				}
				slog.Debug("Found Java class", "name", name, "path", path)
				documents = append(documents, document)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %v: %w", sourceDir, err)
	}
	slog.Info("Found Java class files", "count", len(documents))
	// Index the documents
	err = common.IndexDocuments(db, documents)
	if err != nil {
		return fmt.Errorf("error indexing documents: %w", err)
	}
	return nil
}
