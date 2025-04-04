package javascript

import (
	"log/slog"
	"os"
)

// TODO Call out to acorn instead? https://github.com/acornjs/acorn
// Otto doesn't support ES6+, so it's kind of useless for this.

func DemoASTParser(path string) error {
	jsCode, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	slog.Info("Parsing JavaScript code", "jsCode", string(jsCode))
	return nil
}
