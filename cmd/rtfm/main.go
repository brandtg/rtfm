package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/brandtg/rtfm/internal/java"
)

func ensureOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	outputDir := filepath.Join(home, ".local", "share", "rtfm")
	err = os.MkdirAll(outputDir, 0o755)
	if err != nil {
		panic(err)
	}
	return outputDir
}

func indexJava(args []string) {
	// TODO parse args
	outputDir := ensureOutputDir()
	err := java.Index(outputDir)
	if err != nil {
		slog.Error("Error indexing Java", "error", err)
		panic(err)
	}
}

func findJava(args []string) {
	// Define flags
	flags := flag.NewFlagSet("find", flag.ExitOnError)
	group := flags.String("group", "", "Specify the group ID")
	artifact := flags.String("artifact", "", "Specify the artifact ID")
	version := flags.String("version", "", "Specify the version")
	exact := flags.Bool("exact", false, "Specify exact match")
	format := flags.String(
		"format", "default", "Specify output format (default, class, json, javadoc, source)")
	server := flags.Bool("server", false, "Serve Java classes on http server")
	serverPort := flags.Int("port", 9999, "Port for the HTTP server")
	// Parse flags
	err := flags.Parse(args)
	if err != nil {
		slog.Error("Error parsing flags", "error", err)
		os.Exit(1)
	}
	// Validate arguments
	if len(flags.Args()) < 1 {
		slog.Error("expected pattern")
		os.Exit(1)
	}
	// Find Java classes
	outputDir := ensureOutputDir()
	pattern := flags.Args()[0]
	javaClasses, err := java.Find(outputDir, pattern, *group, *artifact, *version, *exact, *format)
	if err != nil {
		slog.Error("Error finding Java", "error", err)
		panic(err)
	}
	// Serve Java classes on http server
	if *server {
		java.Server(outputDir, javaClasses, *serverPort)
	}
}

func viewJava(args []string) {
	flags := flag.NewFlagSet("view", flag.ExitOnError)
    source := flags.Bool("source", false, "Show source code")
	err := flags.Parse(args)
	if err != nil {
		slog.Error("Error parsing flags", "error", err)
		os.Exit(1)
	}
	flagArgs := flags.Args()
	if len(flagArgs) < 1 {
		slog.Error("expected class name")
		os.Exit(1)
	}
	target := flagArgs[0]
	outputDir := ensureOutputDir()
	_, err = java.View(outputDir, target, *source)
	if err != nil {
		slog.Error("Error viewing Java class", "error", err)
		panic(err)
	}
}

func runJava(args []string) {
	switch args[0] {
	case "index":
		indexJava(args[1:])
	case "find":
		findJava(args[1:])
	case "view":
		viewJava(args[1:])
	default:
		slog.Error("unknown command")
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		slog.Error("expected subcommand")
		os.Exit(1)
	}
	switch os.Args[1] {
    case "java":
        runJava(os.Args[2:])
    default:
        slog.Error("unknown command")
        os.Exit(1)
	}
}
