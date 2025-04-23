/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/brandtg/rtfm/app/golang"
	"github.com/brandtg/rtfm/app/java"
	"github.com/brandtg/rtfm/app/javascript"
	"github.com/brandtg/rtfm/app/python"
	"github.com/spf13/cobra"
)

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index third party libraries",
	Run: func(cmd *cobra.Command, args []string) {
		// Parse arguments
		langName, err := cmd.Flags().GetString("lang")
		if err != nil {
			panic(err)
		}
		// Java
		if langName == "" || langName == "java" {
			err = java.Index()
			if err != nil {
				panic(err)
			}
		}
		// Python
		if langName == "" || langName == "python" {
			err = python.Index()
			if err != nil {
				panic(err)
			}
		}
		// JavaScript
		if langName == "" || langName == "javascript" {
			err = javascript.Index()
			if err != nil {
				panic(err)
			}
		}
		// Go
		if langName == "" || langName == "go" {
			err = golang.Index()
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	indexCmd.Flags().StringP("lang", "l", "", "Language to index")
}
