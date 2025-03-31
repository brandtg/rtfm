/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"

	"github.com/brandtg/rtfm/internal/common"
	"github.com/brandtg/rtfm/internal/java"
	"github.com/spf13/cobra"
)

func runIndex(cmd *cobra.Command, args []string) {
	outputDir := common.EnsureOutputDir()
	err := java.Index(outputDir)
	if err != nil {
		slog.Error("Error indexing Java", "error", err)
	}
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Indexes code and documentation for different programming languages",
	Run:   runIndex,
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
