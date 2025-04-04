package javascript

import (
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
	for _, library := range []string{"jsdoc", "typedoc", "jsdoc-to-markdown"} {
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
