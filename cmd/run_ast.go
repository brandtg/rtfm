package cmd

import (
	"fmt"

	"github.com/brandtg/rtfm/internal/javascript"
	"github.com/spf13/cobra"
)

var runAstCmd = &cobra.Command{
	Use:   "run_ast",
	Short: "A subcommand for AST",
	Run: func(cmd *cobra.Command, args []string) {
		markdown, err := javascript.DemoASTParser(args[0])
		if err != nil {
			panic(err)
		}
		fmt.Println(markdown)
	},

}

func init() {
	rootCmd.AddCommand(runAstCmd)
}
