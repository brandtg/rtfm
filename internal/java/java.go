package java

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

func OutputDir(baseOutputDir string) string {
	return filepath.Join(baseOutputDir, "java")
}

func Index(baseOutputDir string) error {
	outputDir := OutputDir(baseOutputDir)
	repos, err := ListRepos()
	if err != nil {
		slog.Error("Error listing repositories", "error", err)
		return err
	}
	for _, repo := range repos {
		slog.Info("Indexing repo", "repo", repo)
		// Discover artifacts
		artifacts, err := DiscoverArtifacts(repo)
		if err != nil {
			slog.Error("Error discovering artifacts", "repo", repo, "error", err)
			return err
		}
		// Extract artifacts
		for _, artifact := range artifacts {
			_, err = ExtractArtifact(artifact, outputDir)
			if err != nil {
				slog.Error("Error extracting artifact", "artifact", artifact, "error", err)
				return err
			}
		}
		// Connect to SQLite database
		db, err := OpenDB(outputDir)
		if err != nil {
			return err
		}
		defer db.Close()
		err = CreateTables(db)
		if err != nil {
			return err
		}
		// JDK classes
		jdkClasses, err := findJDKClasses()
		if err != nil {
			return err
		}
		err = InsertJavaClasses(db, jdkClasses)
		if err != nil {
			return err
		}
		// Parse class index files to find Java classes
		for _, artifact := range artifacts {
			if artifact.Classifier == "javadoc" {
				path := filepath.Join(outputDir, artifact.OutputDir())
				slog.Debug("Parsing class index file", "artifact", path)
				classIndexFileNames, err := DiscoverClassIndexHtml(path)
				if err != nil {
					return err
				}
				for _, classIndexFileName := range classIndexFileNames {
					javaClasses, err := ParseClassIndexHtml(
						outputDir,
						path, artifact, classIndexFileName)
					if err != nil {
						return err
					}
					err = InsertJavaClasses(db, javaClasses)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func Find(
	baseOutputDir string,
	pattern string,
	group string,
	artifact string,
	version string,
	exact bool,
	format string,
) ([]JavaClass, error) {
	outputDir := OutputDir(baseOutputDir)
	// Connect to SQLite database
	db, err := OpenDB(outputDir)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// Search for Java classes matching pattern
	javaClasses, err := ListJavaClasses(db, pattern, group, artifact, version, exact)
	if err != nil {
		return nil, err
	}
	// Print the results
	switch format {
	case "default":
		for i, javaClass := range javaClasses {
			if i > 0 {
				fmt.Println()
			}
			fmt.Println("Class:")
			fmt.Println("  " + javaClass.Name)
			fmt.Println("Javadoc:")
			fmt.Println("  " + javaClass.Path)
			fmt.Println("Source:")
			fmt.Println("  " + javaClass.Source)
		}
	case "json":
		json, err := json.MarshalIndent(javaClasses, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Println(string(json))
	case "javadoc":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Path)
		}
	case "source":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Source)
		}
	case "class":
		for _, javaClass := range javaClasses {
			fmt.Println(javaClass.Name)
		}
	default:
		slog.Error("Unknown format", "format", format)
	}
	return javaClasses, nil
}

//go:embed templates/index.html
var indexHTML []byte

func setContentType(w http.ResponseWriter, filePath string) {
	switch filepath.Ext(filePath) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
}

func Server(baseOutputDir string, javaClasses []JavaClass, port int) {
	// Set up template
	outputDir := OutputDir(baseOutputDir)
	// Load template
	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"startsWith": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
	}).Parse(string(indexHTML))
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return
	}
	// Create a new HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Title       string
			Description string
			Classes     []JavaClass
		}{
			Title:       "Java Classes",
			Description: "List of Java classes",
			Classes:     javaClasses,
		}
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			slog.Error("Error executing template", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	// Handle /javadoc and look up / serve the file
	http.HandleFunc("/javadoc/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(outputDir, r.URL.Path)
		content, err := os.ReadFile(filePath)
		if err != nil {
			slog.Error("Error reading file", "path", filePath, "error", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		setContentType(w, filePath)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
	// Handle /source and look up / serve the file
	http.HandleFunc("/sources/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(outputDir, r.URL.Path)
		content, err := os.ReadFile(filePath)
		if err != nil {
			slog.Error("Error reading file", "path", filePath, "error", err)
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		setContentType(w, filePath)
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	})
	// Start the server
	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		slog.Info(fmt.Sprintf("Starting server on port http://localhost:%d", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()
	// Wait for Ctrl+C to exit
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	// Shutdown the server
	slog.Info("Shutting down server")
	if err := server.Close(); err != nil {
		slog.Error("Error shutting down server", "error", err)
	}
}
