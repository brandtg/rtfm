package common

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
)

func HighlightCode(code string, language string) (string, error) {
	var buffer bytes.Buffer
	err := quick.Highlight(&buffer, code, language, "terminal256", "monokai")
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
