package promauth

import (
	"strings"
	"unicode"

	"github.com/cprobe/cprobe/lib/fs"
)

func readPasswordFromFile(path string) (string, error) {
	data, err := fs.ReadFileOrHTTP(path)
	if err != nil {
		return "", err
	}
	pass := strings.TrimRightFunc(string(data), unicode.IsSpace)
	return pass, nil
}
