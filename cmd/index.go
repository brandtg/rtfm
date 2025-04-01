/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"

	"github.com/brandtg/rtfm/internal/common"
	"github.com/brandtg/rtfm/internal/java"
	"github.com/brandtg/rtfm/internal/python"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Indexes code and documentation for different programming languages",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

var javaIndexCmd = &cobra.Command{
	Use:   "java",
	Short: "Indexes Java code and documentation",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := common.EnsureOutputDir()
		err := java.Index(outputDir)
		if err != nil {
			slog.Error("Error indexing Java", "error", err)
		}
	},
}

var pythonIndexCmd = &cobra.Command{
	Use:   "python",
	Short: "Indexes Python code and documentation",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := common.EnsureOutputDir()
		err := python.Index(outputDir)
		if err != nil {
			slog.Error("Error indexing Python", "error", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.AddCommand(javaIndexCmd)
	indexCmd.AddCommand(pythonIndexCmd)
}
