package app

import (
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var sgr = regexp.MustCompile("\x1b\\[[0-9;]*m")

// styleMap returns, for each visible byte of the rendered output, the SGR
// parameters active at that point. Two renderings are visually identical when
// their styleMaps and stripped text match, regardless of how many escape
// sequences were emitted.
func styleMap(rendered string) (string, []string) {
	var text strings.Builder
	var states []string
	cur := ""
	i := 0
	for i < len(rendered) {
		if loc := sgr.FindStringIndex(rendered[i:]); loc != nil && loc[0] == 0 {
			seq := rendered[i : i+loc[1]]
			if seq == "\x1b[0m" || seq == "\x1b[m" {
				cur = ""
			} else {
				cur = seq
			}
			i += loc[1]
			continue
		}
		text.WriteByte(rendered[i])
		states = append(states, cur)
		i++
	}
	return text.String(), states
}

// renderPerRune is the pre-optimization rendering, kept here to prove the run
// grouping produces the same visible output.
func renderPerRune(text string, matchBits []bool, isSelected bool, cmdStyle lipgloss.Style) string {
	var b strings.Builder
	for byteIdx, r := range text {
		var charStyle lipgloss.Style
		if byteIdx < len(matchBits) && matchBits[byteIdx] {
			if isSelected {
				charStyle = matchSelectedStyle
			} else {
				charStyle = matchNormalStyle
			}
		} else {
			charStyle = cmdStyle
		}
		b.WriteString(charStyle.Render(string(r)))
	}
	return b.String()
}

func renderRuns(text string, matchBits []bool, isSelected bool, cmdStyle lipgloss.Style) string {
	matchStyle := matchNormalStyle
	if isSelected {
		matchStyle = matchSelectedStyle
	}
	var b strings.Builder
	writeStyledRuns(&b, text, matchBits, matchStyle, cmdStyle)
	return b.String()
}

func TestRenderRunsMatchesPerRune(t *testing.T) {
	texts := []string{
		"git commit -m hello",
		"日本語のコマンド ls -la",
		"",
		"x",
		"ａｂ漢字 mixed ascii",
	}
	idxSets := [][]int{
		nil,
		{0},
		{0, 1, 2},
		{4, 5, 10},
		{0, 3, 6, 9, 12, 15},
		{1, 2, 3, 4, 5, 6, 7, 8},
	}
	for _, text := range texts {
		for _, idxs := range idxSets {
			for _, sel := range []bool{false, true} {
				var matchBits []bool
				if len(idxs) > 0 {
					matchBits = make([]bool, len(text))
					for _, i := range idxs {
						if i >= 0 && i < len(matchBits) {
							matchBits[i] = true
						}
					}
				}
				cmdStyle := itemNormalStyle
				if sel {
					cmdStyle = itemSelectedStyle
				}
				oldOut := renderPerRune(text, matchBits, sel, cmdStyle)
				newOut := renderRuns(text, matchBits, sel, cmdStyle)

				oldText, oldStates := styleMap(oldOut)
				newText, newStates := styleMap(newOut)
				if oldText != newText {
					t.Fatalf("text mismatch for %q %v sel=%v:\n old=%q\n new=%q", text, idxs, sel, oldText, newText)
				}
				if len(oldStates) != len(newStates) {
					t.Fatalf("state count mismatch for %q %v sel=%v", text, idxs, sel)
				}
				for i := range oldStates {
					if oldStates[i] != newStates[i] {
						t.Fatalf("style mismatch at byte %d of %q %v sel=%v: old=%q new=%q",
							i, text, idxs, sel, oldStates[i], newStates[i])
					}
				}
				if ansi.StringWidth(oldOut) != ansi.StringWidth(newOut) {
					t.Fatalf("width mismatch for %q", text)
				}
				if len(newOut) > len(oldOut) {
					t.Fatalf("run output larger than per-rune output for %q", text)
				}
			}
		}
	}
}
