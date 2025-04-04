package javascript

import (
	"log/slog"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

// TODO Call out to acorn instead? https://github.com/acornjs/acorn

func DemoASTParser() {

	// JavaScript code to parse
	jsCode := `
		/**
		 * This is a simple JavaScript function
		 */
		function add(a, b) {
			// This is a simple JavaScript function
			return a + b;
		}

		var x = 10;
	`

	program, err := parser.ParseFile(nil, "example.js", jsCode, 0)
	if err != nil {
		slog.Error("Error parsing JavaScript code", "error", err)
		return
	}

	for _, stmt := range program.Body {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			slog.Info("Function Statement", "name", s.Function.Name)
			slog.Info("Function Parameters", "params", s.Function.ParameterList)
			for _, param := range s.Function.ParameterList.List {
				slog.Info("Parameter", "name", param.Name)
			}
		}
	}
}
