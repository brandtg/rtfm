package cmd

import (
	"github.com/brandtg/rtfm/internal/javascript"
	"github.com/spf13/cobra"
)

var runAstCmd = &cobra.Command{
	Use:   "run_ast",
	Short: "A subcommand for AST",
	Run: func(cmd *cobra.Command, args []string) {
		javascript.DemoASTParser(args[0])
	},

}

func init() {
	rootCmd.AddCommand(runAstCmd)
}
