// Copyright 2025 Greg Brandt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/brandtg/rtfm/app/common"
	"github.com/brandtg/rtfm/app/golang"
	"github.com/brandtg/rtfm/app/java"
	"github.com/brandtg/rtfm/app/javascript"
	"github.com/brandtg/rtfm/app/python"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index third party libraries",
	Run: func(cmd *cobra.Command, args []string) {
		// Parse arguments
		langName, err := cmd.Flags().GetString("lang")
		if err != nil {
			panic(err)
		}
		remove, err := cmd.Flags().GetBool("remove")
		if err != nil {
			panic(err)
		}
		if remove {
			err = common.RemoveOutputDir()
			if err != nil {
				panic(err)
			}
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
	indexCmd.Flags().StringP("lang", "l", "", "Language to index")
	indexCmd.Flags().Bool("remove", false, "Remove any existing index")
}
