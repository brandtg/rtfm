/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/brandtg/rtfm/internal/common"
	"github.com/brandtg/rtfm/internal/java"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find code and documentation",
    Run: func(cmd *cobra.Command, args []string) {
        _ = cmd.Usage()
    },
}

var javaFindCmd = &cobra.Command{
    Use:   "java",
    Short: "A subcommand for find",
    Run: func(cmd *cobra.Command, args []string) {
		group, _ := cmd.Flags().GetString("group")
		artifact, _ := cmd.Flags().GetString("artifact")
		version, _ := cmd.Flags().GetString("version")
		exact, _ := cmd.Flags().GetBool("exact")
        format, _ := cmd.Flags().GetString("format")
        outputDir := common.EnsureOutputDir()
        _, err := java.Find(outputDir, args[0], group, artifact, version, exact, format)
        if err != nil {
            slog.Error("Error finding Java classes", "error", err)
            panic(err)
        }
    },
    Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(findCmd)
    // Java find command
	javaFindCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    javaFindCmd.Flags().StringP("group", "g", "", "Specify the group ID")
    javaFindCmd.Flags().StringP("artifact", "a", "", "Specify the artifact ID")
    javaFindCmd.Flags().StringP("version", "v", "", "Specify the version")
    javaFindCmd.Flags().BoolP("exact", "e", false, "Specify exact match")
    javaFindCmd.Flags().StringP("format", "f", "default", "Specify output format (default, class, json, javadoc, source)")
	findCmd.AddCommand(javaFindCmd)
}
