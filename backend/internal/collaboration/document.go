package collaboration

import (
	"time"
)

// DocumentMetadata represents the descriptive state of an active document.
type DocumentMetadata struct {
	ID        string            `json:"id"`
	Language  SupportedLanguage `json:"language"`
	Title     string            `json:"title"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewDefaultDocument creates initial metadata for a room's primary document.
func NewDefaultDocument(lang SupportedLanguage) DocumentMetadata {
	now := time.Now().UTC()
	return DocumentMetadata{
		ID:        "default",
		Language:  lang,
		Title:     defaultFilenameForLanguage(lang),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func defaultFilenameForLanguage(lang SupportedLanguage) string {
	switch lang {
	case LangC:
		return "main.c"
	case LangCPP:
		return "main.cpp"
	case LangJava:
		return "Main.java"
	case LangPython:
		return "main.py"
	case LangGo:
		return "main.go"
	case LangJavaScript:
		return "index.js"
	case LangTypeScript:
		return "index.ts"
	case LangCSharp:
		return "Program.cs"
	case LangRust:
		return "main.rs"
	default:
		return "main.txt"
	}
}
