package common

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	_ "github.com/mattn/go-sqlite3"
)

type Language int

const (
	Java Language = iota
	Python
	Javascript
)

func LanguageFromName(name string) Language {
	switch strings.ToLower(name) {
	case "java":
		return Java
	case "python":
		return Python
	case "javascript":
		return Javascript
	default:
		return -1
	}
}

func NameFromLanguage(language Language) string {
	switch language {
	case Java:
		return "java"
	case Python:
		return "python"
	case Javascript:
		return "javascript"
	default:
		return ""
	}
}

type SearchDocument struct {
	Language Language
	Name     string
	Path     string
}

func MakeFuzzy(pattern string) string {
	re := regexp.MustCompile(`\s+`)
	return "%" + re.ReplaceAllString(pattern, "%") + "%"
}

func RunFzf(text *bytes.Buffer) (string, error) {
	// Check if fzf is installed
	_, err := exec.LookPath("fzf")
	if err != nil {
		panic(fmt.Errorf("fzf not found in PATH. Please install fzf to use this feature"))
	}
	// Run fzf over the text
	fzf := exec.Command("fzf")
	fzf.Stdin = text
	var selected bytes.Buffer
	fzf.Stdout = &selected
	fzf.Stderr = os.Stderr
	if err := fzf.Run(); err != nil {
		return "", err
	}
	// Get the selected text
	selection := strings.TrimSpace(selected.String())
	return selection, nil
}

func RunFzfSearchDocuments(docs []*SearchDocument) (*SearchDocument, error) {
	// Create a buffer of the document names
	lines := make([]string, len(docs))
	for i, doc := range docs {
		// lines[i] = doc.Name
		lines[i] = strings.Join([]string{NameFromLanguage(doc.Language), doc.Name}, "\t")
	}
	lines = Dedupe(lines)
	slices.Sort(lines)
	text := bytes.NewBufferString(strings.Join(lines, "\n"))
	// Get the selected document name via fzf
	selected, err := RunFzf(text)
	if err != nil {
		return nil, err
	}
	selectedTokens := strings.Split(selected, "\t")
	selectedName := selectedTokens[1]
	// Find the document in the list
	for _, doc := range docs {
		if doc.Name == selectedName {
			return doc, nil
		}
	}
	return nil, fmt.Errorf("document not found: %s", selected)
}

func pathAsComment(language Language, path string) string {
	switch language {
	case Java, Javascript:
		return fmt.Sprintf("// %s", path)
	default:
		return fmt.Sprintf("# %s", path)
	}
}

func HighlightCode(code string, language Language, path string) (string, error) {
	// Prepend the path as a comment
	code = pathAsComment(language, path) + "\n\n" + code
	// Highlight the code
	languageName := NameFromLanguage(language)
	var buffer bytes.Buffer
	err := quick.Highlight(&buffer, code, languageName, "terminal256", "monokai")
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func Dedupe(input []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, item := range input {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func DisplayInPager(text string) error {
	cmd := exec.Command("less")
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Error displaying with pager:", err)
	}
	return nil
}
