package java

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	// "net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Link struct {
	Text  string
	Title string
	Href  string
}

type JavaClass struct {
	Name     string
	Path     string
	Source   string
	Artifact *MavenCoordinates
}

func isCodeLink(title string) bool {
	return strings.Contains(title, "class in") ||
		strings.Contains(title, "interface in")
}

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

func ParseClassIndexHtml(
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

func DiscoverClassIndexHtml(inputDir string) ([]string, error) {
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
