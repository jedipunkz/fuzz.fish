package ui

import "github.com/koki-develop/go-fzf"

// NewFinder creates a new fzf finder with Tokyo Night theme
func NewFinder() (*fzf.FZF, error) {
	return fzf.New(
		fzf.WithStyles(
			fzf.WithStylePrompt(fzf.Style{ForegroundColor: ColorBlue}),
			fzf.WithStyleInputText(fzf.Style{ForegroundColor: ColorForeground}),
			fzf.WithStyleCursor(fzf.Style{ForegroundColor: ColorBlue}),
			// Selection background set to brighter purple
			fzf.WithStyleCursorLine(fzf.Style{ForegroundColor: ColorForeground, BackgroundColor: ColorSelectionBg, Bold: true}),
			fzf.WithStyleMatches(fzf.Style{ForegroundColor: ColorOrange}),
			fzf.WithStyleSelectedPrefix(fzf.Style{ForegroundColor: ColorBlue}),
			fzf.WithStyleUnselectedPrefix(fzf.Style{ForegroundColor: ColorComment}),
		),
		fzf.WithInputPosition(fzf.InputPositionBottom),
	)
}
