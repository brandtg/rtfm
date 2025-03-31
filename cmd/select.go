/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func checkFzfInstalled() error {
    _, err := exec.LookPath("fzf")
    if err != nil {
        return fmt.Errorf("fzf not found in PATH. Please install fzf to use this feature")
    }
    return nil
}


var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Uses fzf to select a file",
	Run: func(cmd *cobra.Command, args []string) {
        // Check if fzf is installed
        err := checkFzfInstalled()
        if err != nil {
            panic(err)
        }
        // Run find command
        findCmd := exec.Command(os.Args[0], append([]string{"find"}, args...)...)
        var findOutput bytes.Buffer
        findCmd.Stdout = &findOutput
        findCmd.Stderr = nil
        if err = findCmd.Run(); err != nil {
            panic(err)
        }
        // Pipe the output to fzf
        fzf := exec.Command("fzf")
        fzf.Stdin = &findOutput
        var selected bytes.Buffer
        fzf.Stdout = &selected
        fzf.Stderr = os.Stderr
        if err = fzf.Run(); err != nil {
            panic(err)
        }
        selection := strings.TrimSpace(selected.String())
        // View the selected file
        language := args[0]
		source, _ := cmd.Flags().GetBool("source")
        viewArgs := []string{"view", language, selection}
        if source {
            viewArgs = append(viewArgs, "--source")
        }
        viewCmd := exec.Command(os.Args[0], viewArgs...)
        viewCmd.Stdout = os.Stdout
        viewCmd.Stderr = os.Stderr
        if err = viewCmd.Run(); err != nil {
            panic(err)
        }
	},
    Args: cobra.MinimumNArgs(2),
}

func init() {
	rootCmd.AddCommand(selectCmd)
    selectCmd.Flags().BoolP("source", "s", false, "View sources")
}
