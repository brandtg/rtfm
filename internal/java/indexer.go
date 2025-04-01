package java

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/brandtg/rtfm/internal/common"
)

func isLocalLink(href string) bool {
	return !strings.Contains(href, "http://") &&
		!strings.Contains(href, "https://") &&
		!strings.Contains(href, "is-external=true")
}

func findClassLinks(reader io.Reader) ([]Link, error) {
	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	// Find all links in the document
	var links []Link
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, hasHref := s.Attr("href")
		if hasHref {
			title, hasTitle := s.Attr("title")
			if hasTitle && isCodeLink(title) && isLocalLink(href) {
				text := s.Text()
				links = append(links, Link{Text: text, Title: title, Href: href})
			}
		}
	})
	return links, nil
}

func parseJavaClass(
	outputDir string,
	basePath string,
	artifact *MavenCoordinates,
	javaDocPath string,
) JavaClass {
	classPath := strings.Replace(javaDocPath, basePath, "", 1)[1:]
	className := strings.ReplaceAll(classPath, ".html", "")
	className = strings.ReplaceAll(className, "/", ".")
	javaDocPath = strings.Replace(javaDocPath, outputDir, "", 1)[1:]
	source := strings.ReplaceAll(javaDocPath, "javadoc", "sources")
	source = strings.ReplaceAll(source, ".html", ".java")
	return JavaClass{
		Name:     className,
		Path:     javaDocPath,
		Artifact: artifact,
		Source:   source,
	}
}

func parseClassIndexHtml(
	outputDir string,
	basePath string,
	artifact *MavenCoordinates,
	classIndexPath string,
) ([]JavaClass, error) {
	// Open the class index HTML file
	file, err := os.Open(classIndexPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// Find the links in the HTML
	links, err := findClassLinks(file)
	if err != nil {
		return nil, err
	}
	// Parse the links into JavaClass objects
	javaClasses := make([]JavaClass, len(links))
	for i, link := range links {
		javaClasses[i] = parseJavaClass(
			outputDir,
			basePath,
			artifact,
			filepath.Clean(filepath.Join(filepath.Dir(classIndexPath), link.Href)))
	}
	return javaClasses, nil
}

func Index(baseOutputDir string) error {
	// Connect to SQLite database
	outputDir := javaOutputDir(baseOutputDir)
	db, err := common.OpenDB(outputDir, DB_NAME)
	if err != nil {
		return err
	}
	defer db.Close()
	err = createTables(db)
	if err != nil {
		return err
	}
	// Find Java artifact repos
	repos, err := listRepos()
	if err != nil {
		slog.Error("Error listing repositories", "error", err)
		return err
	}
	for _, repo := range repos {
		slog.Info("Indexing repo", "repo", repo)
		// Discover artifacts
		artifacts, err := discoverArtifacts(repo)
		if err != nil {
			slog.Error("Error discovering artifacts", "repo", repo, "error", err)
			return err
		}
		// Extract artifacts
		for _, artifact := range artifacts {
			_, err = extractArtifact(artifact, outputDir)
			if err != nil {
				slog.Error("Error extracting artifact", "artifact", artifact, "error", err)
				return err
			}
		}
		// JDK classes
		jdkClasses, err := findJDKClasses()
		if err != nil {
			return err
		}
		err = insertJavaClasses(db, jdkClasses)
		if err != nil {
			return err
		}
		// Parse class index files to find Java classes
		for _, artifact := range artifacts {
			if artifact.Classifier == "javadoc" {
				path := filepath.Join(outputDir, artifact.OutputDir())
				slog.Debug("Parsing class index file", "artifact", path)
				classIndexFileNames, err := discoverClassIndexHtml(path)
				if err != nil {
					return err
				}
				for _, classIndexFileName := range classIndexFileNames {
					javaClasses, err := parseClassIndexHtml(
						outputDir,
						path, artifact, classIndexFileName)
					if err != nil {
						return err
					}
					err = insertJavaClasses(db, javaClasses)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func isCodeLink(title string) bool {
	return strings.Contains(title, "class in") ||
		strings.Contains(title, "interface in")
}

func discoverClassIndexHtml(inputDir string) ([]string, error) {
	// Define the class index files to look for
	classIndexFileNames := map[string]bool{
		"allclasses-frame.html":   true,
		"allclasses-noframe.html": true,
		"package-tree.html":       true,
	}
	// Walk through the input directory
	acc := []string{}
	err := filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			name := filepath.Base(path)
			if classIndexFileNames[name] {
				acc = append(acc, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return acc, nil
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
	javaVersion := match[1]
	slog.Info("Java version found", "version", javaVersion)
	return javaVersion, nil
}

func jdkClassUrl(link Link) string {
	tokens := strings.Split(link.Href, "/")
	prefix := tokens[0]
	suffix := tokens[1:]
	path := append([]string{prefix, "share", "classes"}, suffix...)
	url := strings.Join(path, "/")
	url = "https://github.com/openjdk/jdk/blob/master/src/" + strings.ReplaceAll(url, ".html", ".java")
	return url
}

func findJDKClasses() ([]JavaClass, error) {
	// Create JDK url based on Java version
	javaVersion, err := findJavaVersion()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://docs.oracle.com/en/java/javase/%s/docs/api/", javaVersion)
	// Fetch the JDK classes from the URL
	slog.Info("Fetching JDK classes from URL", "url", url)
	res, err := http.Get(url + "allclasses-index.html")
	if err != nil {
		slog.Error("Error fetching URL", "url", url, "error", err)
		return nil, err
	}
	defer res.Body.Close()
	// Extract class and interface links from the response
	links, err := findClassLinks(res.Body)
	if err != nil {
		return nil, err
	}
	// Parse the links into JavaClass objects
	javaClasses := make([]JavaClass, len(links))
	for i, link := range links {
		tokens := strings.Split(link.Href, "/")
		className := strings.Join(tokens[1:], ".")
		className = strings.ReplaceAll(className, ".html", "")
		javaClasses[i] = JavaClass{
			Name:     className,
			Path:     url + link.Href,
			Source:   jdkClassUrl(link),
			Artifact: &MavenCoordinates{},
		}
	}
	return javaClasses, nil
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

func extractArtifact(coords *MavenCoordinates, outputDir string) (bool, error) {
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

func listRepos() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	// TODO Add gradle
	mavenDir := filepath.Join(home, ".m2", "repository")
	return []string{mavenDir}, nil
}
