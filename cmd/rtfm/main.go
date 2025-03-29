package main

import (
	// "flag"
	"github.com/brandtg/rtfm/internal/core"
	"log/slog"
	"os"
)

func index(args []string) {
	// TODO parse args
	core.Index()
}

func find(args []string) {
	// TODO parse args
	core.Find()
}

func main() {
	if len(os.Args) < 2 {
		slog.Error("expected subcommand")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "index":
		index(os.Args[2:])
	case "find":
		find(os.Args[2:])
	default:
		slog.Error("unknown command")
		os.Exit(1)
	}
}
