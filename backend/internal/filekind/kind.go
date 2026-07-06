package filekind

import (
	"errors"
	"strings"
)

func IsMarkdown(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".md")
}

func IsCode(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, ".go") || strings.HasSuffix(n, ".py")
}

func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if IsMarkdown(name) || IsCode(name) {
		return nil
	}
	return errors.New("file name must end with .go, .py, or .md")
}
