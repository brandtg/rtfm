package javascript

import (
	"encoding/json"
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
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}
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
	slog.Info("Parsed AST", "ast", ast)
	// jsCode, err := os.ReadFile(path)
	// if err != nil {
	// 	return err
	// }
	// slog.Info("Parsing JavaScript code", "jsCode", string(jsCode))
	// npx --yes --package=@babel/parser node parse-ast.js ~/Desktop/test.js

	return nil
}
