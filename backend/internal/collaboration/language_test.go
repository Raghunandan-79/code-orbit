package collaboration

import (
	"testing"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected SupportedLanguage
		hasErr   bool
	}{
		{"", LangJavaScript, false},
		{"c", LangC, false},
		{"cpp", LangCPP, false},
		{"C++", LangCPP, false},
		{"java", LangJava, false},
		{"python", LangPython, false},
		{"PY", LangPython, false},
		{"go", LangGo, false},
		{"golang", LangGo, false},
		{"javascript", LangJavaScript, false},
		{"JS", LangJavaScript, false},
		{"typescript", LangTypeScript, false},
		{"TS", LangTypeScript, false},
		{"csharp", LangCSharp, false},
		{"c#", LangCSharp, false},
		{"rust", LangRust, false},
		{"RS", LangRust, false},
		{"ruby", "", true},
		{"php", "", true},
		{"brainfuck", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res, err := NormalizeLanguage(tt.input)
			if (err != nil) != tt.hasErr {
				t.Fatalf("input %q: expected error %v, got %v", tt.input, tt.hasErr, err)
			}
			if res != tt.expected {
				t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, res)
			}
		})
	}
}
