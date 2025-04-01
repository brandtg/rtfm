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

func checkFzfInstalled() {
	_, err := exec.LookPath("fzf")
	if err != nil {
		panic(fmt.Errorf("fzf not found in PATH. Please install fzf to use this feature"))
	}
}

func runFindCommand(args []string) *bytes.Buffer {
	findCmd := exec.Command(os.Args[0], append([]string{"find"}, args...)...)
	var findOutput bytes.Buffer
	findCmd.Stdout = &findOutput
	findCmd.Stderr = nil
	if err := findCmd.Run(); err != nil {
		panic(err)
	}
	return &findOutput
}

func runFzf(findOutput *bytes.Buffer) string {
	fzf := exec.Command("fzf")
	fzf.Stdin = findOutput
	var selected bytes.Buffer
	fzf.Stdout = &selected
	fzf.Stderr = os.Stderr
	if err := fzf.Run(); err != nil {
		panic(err)
	}
	selection := strings.TrimSpace(selected.String())
	return selection
}

func runViewCommand(args []string) {
	viewCmd := exec.Command(os.Args[0], args...)
	viewCmd.Stdout = os.Stdout
	viewCmd.Stderr = os.Stderr
	if err := viewCmd.Run(); err != nil {
		panic(err)
	}
}

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Uses fzf to select a file",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Usage()
	},
	Args: cobra.MinimumNArgs(2),
}

var javaSelectCmd = &cobra.Command{
	Use:   "java",
	Short: "A subcommand for select",
	Run: func(cmd *cobra.Command, args []string) {
		checkFzfInstalled()
		findOutput := runFindCommand(append([]string{"java"}, args...))
		selection := runFzf(findOutput)
		viewArgs := []string{"view", "java", selection}
		source, _ := cmd.Flags().GetBool("source")
		if source {
			viewArgs = append(viewArgs, "--source")
		}
        runViewCommand(viewArgs)
	},
	Args: cobra.MinimumNArgs(1),
}

var pythonSelectCmd = &cobra.Command{
	Use:   "python",
	Short: "A subcommand for select",
	Run: func(cmd *cobra.Command, args []string) {
		checkFzfInstalled()
        findArgs := []string{"python"}
        findArgs = append(findArgs, args...)
        venv, _ := cmd.Flags().GetString("venv")
        if venv != "" {
            findArgs = append(findArgs, "--venv", venv)
        }
		findOutput := runFindCommand(findArgs)
		selection := runFzf(findOutput)
		viewArgs := []string{"view", "python", selection}
		source, _ := cmd.Flags().GetBool("source")
		if source {
			viewArgs = append(viewArgs, "--source")
		}
        runViewCommand(viewArgs)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	selectCmd.AddCommand(javaSelectCmd)
	javaSelectCmd.Flags().BoolP("source", "s", false, "View sources")
    selectCmd.AddCommand(pythonSelectCmd)
    pythonSelectCmd.Flags().BoolP("source", "s", false, "View sources")
    pythonSelectCmd.Flags().StringP("venv", "v", "", "Python virtual environment")
	rootCmd.AddCommand(selectCmd)
}
