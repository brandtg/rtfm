package javascript

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

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
