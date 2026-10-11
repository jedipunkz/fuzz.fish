package ui

import (
	"strings"
	"testing"
)

func TestHighlightCode_SupportedLanguage(t *testing.T) {
	code := "func main() {\n\tfmt.Println(\"hello\")\n}\n"

	got, err := HighlightCode(code, "main.go")
	if err != nil {
		t.Fatal(err)
	}
	if got == code {
		t.Fatalf("output = plain input, want highlighted")
	}
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("output %q has no ANSI escape sequences", got)
	}
	if !strings.Contains(got, "Println") {
		t.Errorf("output %q lost the source content", got)
	}
}

func TestHighlightCode_UnsupportedLanguage(t *testing.T) {
	code := "def meth\n  :sym\nend\n"

	got, err := HighlightCode(code, "script.rb")
	if err != nil {
		t.Fatal(err)
	}
	if got != code {
		t.Errorf("HighlightCode = %q, want input unchanged for unsupported language", got)
	}
}

func TestHighlightCode_NoLexer(t *testing.T) {
	code := "some ordinary words\n"

	got, err := HighlightCode(code, "notes")
	if err != nil {
		t.Fatal(err)
	}
	if got != code {
		t.Errorf("HighlightCode = %q, want input unchanged when no lexer matches", got)
	}
}

func TestHighlightCode_EmptyContent(t *testing.T) {
	got, err := HighlightCode("", "main.go")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("HighlightCode(\"\", \"main.go\") = %q, want empty", got)
	}
}
