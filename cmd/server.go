/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/brandtg/rtfm/internal/common"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO Improve this with search etc.
		port, _ := cmd.Flags().GetInt("port")
		dir := common.EnsureOutputDir()
		address := fmt.Sprintf(":%d", port)
		fmt.Printf("Serving files from %s on http://localhost:%d\n", dir, port)
		fs := http.FileServer(http.Dir(dir))
		http.Handle("/", http.StripPrefix("/", fs)) // Strip prefix for cleaner URLs
		if err := http.ListenAndServe(address, nil); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
}
