package utils

import (
	"bytes"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Tokyo Night colors for syntax highlighting
var _ = styles.Register(chroma.MustNewStyle("tokyonight", chroma.StyleEntries{
	chroma.Comment:           "#9aa5ce italic",
	chroma.CommentPreproc:    "#7dcfff",
	chroma.Keyword:           "#bb9af7 bold",
	chroma.KeywordNamespace:  "#bb9af7",
	chroma.KeywordType:       "#2ac3de",
	chroma.Operator:          "#89ddff",
	chroma.Punctuation:       "#c0caf5",
	chroma.Name:              "#c0caf5",
	chroma.NameBuiltin:       "#2ac3de",
	chroma.NameClass:         "#7dcfff",
	chroma.NameFunction:      "#7dcfff",
	chroma.NameVariable:      "#c0caf5",
	chroma.String:            "#e0af68",
	chroma.StringEscape:      "#bb9af7",
	chroma.Number:            "#ff9e64",
	chroma.Generic:           "#c0caf5",
	chroma.GenericDeleted:    "#f7768e",
	chroma.GenericEmph:       "italic",
	chroma.GenericInserted:   "#9ece6a",
	chroma.GenericStrong:     "bold",
	chroma.GenericSubheading: "#7dcfff",
	chroma.Background:        " bg:#1a1b26",
}))

// HighlightCode performs syntax highlighting on code
func HighlightCode(code string, filename string) (string, error) {
	// Determine lexer from filename
	lexer := lexers.Match(filename)
	if lexer == nil {
		// Try to analyse the content
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		// Fallback to plain text
		return code, nil
	}

	// Only support specific languages
	lexerName := lexer.Config().Name
	supportedLanguages := map[string]bool{
		"Go":         true,
		"Python":     true,
		"JavaScript": true,
		"TypeScript": true,
		"JSON":       true,
		"Rust":       true,
		"YAML":       true,
	}

	if !supportedLanguages[lexerName] {
		// Not in supported list, return plain text
		return code, nil
	}

	lexer = chroma.Coalesce(lexer)

	// Use Tokyo Night style
	style := styles.Get("tokyonight")
	if style == nil {
		style = styles.Fallback
	}

	// Create terminal formatter with 256 colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	// Format output
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}
