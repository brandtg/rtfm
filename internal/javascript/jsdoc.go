package javascript

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

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

func parseAST(
	installDir string,
	path string,
) (*map[string]any, error) {
	cmd := exec.Command("npx", "--yes", "--package=@babel/parser", "node", "parse-ast.js", path)
	cmd.Dir = installDir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(output))
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	/*

➜  rtfm git:(brandtg/jsdoc-ast) ✗ cat ~/Desktop/test.js

var myFunctionVariable = () => console.log("hello world");

function myFunction() {
    console.log("Hello world");
}

	.program.body[] appears to be the list of AST nodes

➜  rtfm git:(brandtg/jsdoc-ast) ✗ jq '.program.body[].type' tmp/ast.json
"VariableDeclaration"
"VariableDeclaration"
"VariableDeclaration"
"VariableDeclaration"
"ExpressionStatement"

➜  rtfm git:(brandtg/jsdoc-ast) ✗ jq '.program.body[].kind' tmp/ast.json
"const"
"const"
"const"
"const"
null

Some of this is comming from that wrapper script

	const parser = require("@babel/parser");
	const fs = require("fs");
	const code = fs.readFileSync(process.argv[1], "utf8");
	const ast = parser.parse(code, {
		sourceType: "unambiguous",
		plugins: ["jsx", "typescript", "classProperties", "decorators-legacy"],
	});
	console.log(JSON.stringify(ast, null, 2));

	*/

	// value, ok := result["keyName"].(string)
	// if !ok {
	//     // Handle the case where the value is not a string
	//     return fmt.Errorf("keyName is not a string")
	// }
	// fmt.Println("Value:", value)

	return &result, nil
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
	slog.Debug("Parsed AST", "ast", ast)
	// jsCode, err := os.ReadFile(path)
	// if err != nil {
	// 	return err
	// }
	// slog.Info("Parsing JavaScript code", "jsCode", string(jsCode))
	// npx --yes --package=@babel/parser node parse-ast.js ~/Desktop/test.js

	return nil
}
