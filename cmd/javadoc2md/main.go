package main

import (
	"fmt"
	"os"

	"github.com/brandtg/rtfm/internal/java"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: filename")
		os.Exit(1)
	}
	filename := os.Args[1]
	markdown := java.FormatMarkdown(filename)
	fmt.Println(markdown)
}
