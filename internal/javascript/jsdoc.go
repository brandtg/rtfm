package javascript

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/brandtg/rtfm/internal/common"
)

// Other libraries to consider:
// https://typedoc.org/
// https://github.com/jsdoc2md/jsdoc-to-markdown/
// https://jsdoc.app/about-commandline
// https://tsdoc.org/

func installNpmLibrary(installDir string, library string) error {
	cmd := exec.Command("npm", "install", "--save", library)
	cmd.Dir = installDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install typedoc: %w", err)
	}
	slog.Debug("Installed library", "library", library, "installDir", installDir)
	return nil
}

func checkInstall() (string, error) {
	// Create directory for node project
	installDir := filepath.Join(common.EnsureOutputDir(), "javascript", "jsdoc")
	if common.Exists(installDir) {
		slog.Debug("JSDoc already installed", "installDir", installDir)
		return installDir, nil
	}
	// Initialize the node project
	os.MkdirAll(installDir, os.ModePerm)
	cmd := exec.Command("npm", "init", "-y")
	cmd.Dir = installDir
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to initialize node project: %w", err)
	}
	slog.Debug("Initialized node project", "installDir", installDir)
	// Install libraries
	for _, library := range []string{"jsdoc", "typedoc", "jsdoc-to-markdown", "esprima", "@babel/parser"} {
		err := installNpmLibrary(installDir, library)
		if err != nil {
			return "", fmt.Errorf("failed to install %s: %w", library, err)
		}
	}
	// Write an empty configuration file for jsdoc
	configFilePath := filepath.Join(installDir, "empty-config.json")
	err := os.WriteFile(configFilePath, []byte("{}"), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write jsdoc config file: %w", err)
	}
	slog.Debug("Created empty jsdoc config file", "path", configFilePath)
	// Write the script to parse AST using @babel/parser
	script := `
	const parser = require("@babel/parser");
	const fs = require("fs");
	const code = fs.readFileSync(process.argv[2], "utf8");
	const ast = parser.parse(code, {
		sourceType: "unambiguous",
		plugins: ["jsx", "typescript", "classProperties", "decorators-legacy"],
	});
	console.log(JSON.stringify(ast, null, 2));
	`
	scriptFilePath := filepath.Join(installDir, "parse-ast.js")
	err = os.WriteFile(scriptFilePath, []byte(script), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write jsdoc script file: %w", err)
	}
	return installDir, nil
}

func jsdocMarkdown(installDir string, target string) (string, error) {
	// TODO This generates the default markdown, but we can provide --json to get the structured
	// data that's provided to the template, in order to render something simpler / more appropriate
	// for the terminal.
	cmd := exec.Command(
		"node", "node_modules/jsdoc-to-markdown/bin/cli.js", target, "-c", "empty-config.json")
	cmd.Dir = installDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run jsdoc2md: %w: %s", err, string(output))
	}
	doc := string(output)
	return doc, nil
}

func ParseJSDoc(target string) (string, error) {
	installDir, err := checkInstall()
	if err != nil {
		return "", err
	}
	doc, err := jsdocMarkdown(installDir, target)
	if err != nil {
		return "", err
	}
	return doc, nil
}

type ASTFunction struct {
	Name	   string
	Parameters []string
}

func parseASTFunction(node map[string]any) (*ASTFunction, error) {
	name, ok := node["id"].(map[string]any)["name"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to parse function name")
	}
	paramNodes, ok := node["params"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse function parameters")
	}
	params := make([]string, 0)
	for _, paramNode := range paramNodes {
		param, ok := paramNode.(map[string]any)["name"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse parameter name")
		}
		if param != "" {
			params = append(params, param)
		}
	}
	return &ASTFunction{
		Name: name,
		Parameters: params,
	}, nil
}

type ASTVariable struct {
	Kind string
	Name string
}

func parseASTVariables(node map[string]any) ([]*ASTVariable, error) {
	declarations, ok := node["declarations"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse variable declarations")
	}
	variables := make([]*ASTVariable, 0)
	for _, declaration := range declarations {
		name, ok := declaration.(map[string]any)["id"].(map[string]any)["name"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse variable name")
		}
		kind, ok := node["kind"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse variable kind")
		}
		if name != "" && kind != "" {
			variables = append(variables, &ASTVariable{
				Name: name,
				Kind: kind,
			})
		}
	}
	return variables, nil
}

type AST struct {
	Functions []*ASTFunction
	Variables []*ASTVariable
}

func parseAST(
	installDir string,
	path string,
) (*AST, error) {
	// Use babel to compute the AST
	cmd := exec.Command("npx", "--yes", "--package=@babel/parser", "node", "parse-ast.js", path)
	cmd.Dir = installDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	// Parse output as JSON
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}
	// Find all the functions and variables in the AST
	allASTFunctions := make([]*ASTFunction, 0)
	allASTVariables := make([]*ASTVariable, 0)
	body, ok := result["program"].(map[string]any)["body"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to parse AST body")
	}
	for _, node := range body {
		switch node.(map[string]any)["type"] {
		case "FunctionDeclaration":
			astFunction, err := parseASTFunction(node.(map[string]any))
			if err != nil {
				return nil, err
			}
			allASTFunctions = append(allASTFunctions, astFunction)
		case "VariableDeclaration":
			astVariables, err := parseASTVariables(node.(map[string]any))
			if err != nil {
				return nil, err
			}
			allASTVariables = append(allASTVariables, astVariables...)
		}
	}
	return &AST{
		Functions: allASTFunctions,
		Variables: allASTVariables,
	}, nil
}

func makeASTMarkdown(path string, ast *AST) string {
	var builder strings.Builder
	// Path
	builder.WriteString(path + "\n")
	builder.WriteString(strings.Repeat("=", len(path)) + "\n\n")
	// Functions
	builder.WriteString("Functions\n")
	builder.WriteString(strings.Repeat("-", len("Functions")) + "\n")
	for _, function := range ast.Functions {
		builder.WriteString(fmt.Sprintf("%s(", function.Name))
		if len(function.Parameters) > 0 {
			builder.WriteString(strings.Join(function.Parameters, ", "))
		}
		builder.WriteString(")\n")
	}
	// Variables
	builder.WriteString("\nVariables\n")
	builder.WriteString(strings.Repeat("-", len("Variables")) + "\n")
	for _, variable := range ast.Variables {
		builder.WriteString(fmt.Sprintf("%s %s\n", variable.Kind, variable.Name))
	}
	return builder.String()
}

func DemoASTParser(
	path string,
) error {
	installDir, err := checkInstall()
	if err != nil {
		return err
	}
	ast, err := parseAST(installDir, path)
	if err != nil {
		return err
	}

	// for _, variable := range ast.Variables {
	// 	slog.Info("Variable", "name", variable.Name, "kind", variable.Kind)
	// }
	// for _, function := range ast.Functions {
	// 	slog.Info("Function", "name", function.Name, "parameters", function.Parameters)
	// 	for _, param := range function.Parameters {
	// 		slog.Info("Parameter", "name", param)
	// 	}
	// }
	markdown := makeASTMarkdown(path, ast)
	fmt.Println(markdown)

	return nil
}
