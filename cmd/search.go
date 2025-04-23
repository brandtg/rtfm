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
	"os"
	"os/exec"

	"github.com/brandtg/rtfm/app/common"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for code snippets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse arguments
		query := args[0]
		langName, err := cmd.Flags().GetString("lang")
		if err != nil {
			panic(err)
		}
		lang := common.LanguageFromName(langName)
		exact, err := cmd.Flags().GetBool("exact")
		if err != nil {
			panic(err)
		}
		// Open the database
		db, err := common.OpenDB()
		if err != nil {
			panic(err)
		}
		defer db.Close()
		// Search for code snippets
		docs, err := common.FindDocuments(db, lang, query, exact)
		if err != nil {
			panic(err)
		}
		// Interactive loop to select and view code files
		var filterQuery string
		var selected *common.SearchDocument
		for {
			// Select the code by name
			filterQuery, selected, err = common.RunFzfSearchDocuments(filterQuery, docs)
			if err != nil {
				// If fzf was closed just exit cleanly
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
					break
				}
				panic(err)
			}
			// Read the code from the file
			code, err := os.ReadFile(selected.Path)
			if err != nil {
				panic(err)
			}
			// Highlight the code
			highlightedCode, err := common.HighlightCode(
				string(code),
				selected.Language,
				selected.Path,
			)
			if err != nil {
				panic(err)
			}
			// Display the code in a pager
			common.DisplayInPager(highlightedCode)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("lang", "l", "", "Language to search for")
	searchCmd.Flags().BoolP("exact", "e", false, "Exact match")
}
