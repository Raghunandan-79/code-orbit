package collaboration

import (
	"fmt"
	"strings"
)

// SupportedLanguage represents a validated language identifier.
type SupportedLanguage string

const (
	LangC          SupportedLanguage = "c"
	LangCPP        SupportedLanguage = "cpp"
	LangJava       SupportedLanguage = "java"
	LangPython     SupportedLanguage = "python"
	LangGo         SupportedLanguage = "go"
	LangJavaScript SupportedLanguage = "javascript"
	LangTypeScript SupportedLanguage = "typescript"
	LangCSharp     SupportedLanguage = "csharp"
	LangRust       SupportedLanguage = "rust"
)

var validLanguages = map[string]SupportedLanguage{
	"c":          LangC,
	"cpp":        LangCPP,
	"c++":        LangCPP,
	"java":       LangJava,
	"python":     LangPython,
	"py":         LangPython,
	"go":         LangGo,
	"golang":     LangGo,
	"javascript": LangJavaScript,
	"js":         LangJavaScript,
	"typescript": LangTypeScript,
	"ts":         LangTypeScript,
	"csharp":     LangCSharp,
	"c#":         LangCSharp,
	"rust":       LangRust,
	"rs":         LangRust,
}

// NormalizeLanguage validates and returns canonical language string.
// If input is empty, defaults to "javascript".
func NormalizeLanguage(input string) (SupportedLanguage, error) {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	if trimmed == "" {
		return LangJavaScript, nil
	}

	lang, ok := validLanguages[trimmed]
	if !ok {
		return "", fmt.Errorf("unsupported language: %q (supported: c, cpp, java, python, go, javascript, typescript, csharp, rust)", input)
	}

	return lang, nil
}

// IsValid checks if language string is supported.
func IsValid(input string) bool {
	_, err := NormalizeLanguage(input)
	return err == nil
}
