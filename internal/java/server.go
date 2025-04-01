package java

import (
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
)

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
	outputDir := javaOutputDir(baseOutputDir)
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
