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

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
}

var javaViewCmd = &cobra.Command{
	Use:   "java",
	Short: "A subcommand for view",
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetBool("source")
		outputDir := common.EnsureOutputDir()
		_, err := java.View(outputDir, args[0], source)
		if err != nil {
			slog.Error("Error viewing Java classes", "error", err)
			panic(err)
		}
	},
	Args: cobra.ExactArgs(1),
}

var pythonViewCmd = &cobra.Command{
	Use:   "python",
	Short: "A subcommand for view",
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetBool("source")
		outputDir := common.EnsureOutputDir()
		_, err := python.View(outputDir, args[0], source)
		if err != nil {
			slog.Error("Error viewing Python modules", "error", err)
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	javaViewCmd.Flags().BoolP("source", "s", false, "Show source code")
	viewCmd.AddCommand(javaViewCmd)
	pythonViewCmd.Flags().BoolP("source", "s", false, "Show source code")
	viewCmd.AddCommand(pythonViewCmd)
}
