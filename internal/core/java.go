package core

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/fs"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

type Link struct {
	title string
	href  string
}

type JavaClass struct {
	name string
	path string
	jar  string
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

func javaClassFromLink(baseUrl string, link Link) JavaClass {
	name := strings.ReplaceAll(link.href, ".html", "")
	name = strings.ReplaceAll(name, "/", ".")
	return JavaClass{
		name: name,
		path: baseUrl + link.href,
		jar:  "JDK",
	}
}

func findJDKClasses() ([]JavaClass, error) {
    // Create JDK url based on Java version
	javaVersion, err := findJavaVersion()
	if err != nil {
		slog.Error("Error finding Java version", "error", err)
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
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		slog.Error("Error parsing HTML", "url", url, "error", err)
		return nil, err
	}
	slog.Info("Parsed document", "title", doc.Find("title").Text())
	links := []Link{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, hasHref := s.Attr("href")
		if hasHref {
			title, hasTitle := s.Attr("title")
			if hasTitle && (strings.Contains(title, "class in") || strings.Contains(title, "interface in")) {
				links = append(links, Link{title, href})
			}
		}
	})
    // Parse links into java classes
	slog.Info("Extracted links", "count", len(links))
	javaClasses := []JavaClass{}
	for _, link := range links {
		javaClasses = append(javaClasses, javaClassFromLink(url, link))
	}
	return javaClasses, nil
}

func findLocalRepositoryClasses(outputDir string) ([]JavaClass, error) {
    javaClasses := []JavaClass{}
    err := filepath.WalkDir(
        filepath.Join(outputDir, "maven", "javadoc"),
        func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            if !d.IsDir() {
                if strings.HasSuffix(path, ".html") {
                    slog.Debug("Found local repository class", "path", path)
                }
            }
            return nil
        },
    )
    if err != nil {
        slog.Error("Error walking directory", "path", outputDir, "error", err)
        return nil, err
    }
    slog.Info("Found local repository classes", "count", len(javaClasses))
    return javaClasses, nil
}

func createJavaDatabase(outputDir string, javaClasses []JavaClass) error {
	// Create sqlite database
	dbPath := filepath.Join(outputDir, "java.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()
	// Create table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS java_class (
		name TEXT PRIMARY KEY,
        path TEXT,
        jar TEXT
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	slog.Info("Database and table created successfully", "path", dbPath)
	// Insert Java classes into the database
	for _, javaClass := range javaClasses {
		_, err := db.Exec(
			`INSERT OR IGNORE INTO java_class (name, path, jar) VALUES (?, ?, ?)`,
			javaClass.name,
			javaClass.path,
			javaClass.jar,
		)
		if err != nil {
			return err
		}
	}
	slog.Info("Java classes inserted into database", "count", len(javaClasses))
	return nil
}

func indexJava(outputDir string) error {
    // Process javadoc and source artifacts
	_, err := findMavenArtifacts(filepath.Join(outputDir, "maven"))
	if err != nil {
		return fmt.Errorf("error finding Maven artifacts: %w", err)
	}
    // Find JDK classes
	jdkClasses, err := findJDKClasses()
	if err != nil {
		return fmt.Errorf("error finding JDK classes: %w", err)
	}
    // Find local repository classes
    localClasses, err := findLocalRepositoryClasses(outputDir)
    if err != nil {
        return fmt.Errorf("error finding local repository classes: %w", err)
    }
    // Write Java classes to database
    javaClasses := append(jdkClasses, localClasses...)
    err = createJavaDatabase(outputDir, javaClasses)
    if err != nil {
        return fmt.Errorf("error creating Java database: %w", err)
    }
	return nil
}
