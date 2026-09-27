package git

import (
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// branchRecentCommits is how many `git log --oneline` style commits the
// branch preview lists.
const branchRecentCommits = 5

// GeneratePreview generates a preview of the branch: metadata plus the
// branch's last commits, newest first. The walk is anchored on the branch's
// full commit hash; branches without a readable commit keep the metadata-only
// preview.
func (b Branch) GeneratePreview(repoPath string, width, height int) string {
	var sb strings.Builder

	// Branch info
	sb.WriteString(ui.LabelStyle.Render("Branch") + "\n")
	sb.WriteString(ui.ContentStyle.Render(b.Name) + "\n\n")

	// Commit hash
	sb.WriteString(ui.LabelStyle.Render("Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(b.LastCommit) + "\n\n")

	// Type
	sb.WriteString(ui.LabelStyle.Render("Type") + "\n")
	if b.IsCurrent {
		sb.WriteString(ui.ContentStyle.Render("Current branch") + "\n")
	} else if b.IsRemote {
		sb.WriteString(ui.ContentStyle.Render("Remote branch") + "\n")
	} else {
		sb.WriteString(ui.ContentStyle.Render("Local branch") + "\n")
	}

	repo := Repository{Path: repoPath}
	if commits := repo.RecentCommits(b.Hash, branchRecentCommits); len(commits) > 0 {
		sb.WriteString("\n")
		sb.WriteString(ui.ContextHeaderStyle.Render("Recent commits") + "\n")
		for _, c := range commits {
			line := c.Hash + " " + strings.ReplaceAll(c.Subject, "\n", " ")
			sb.WriteString(ui.ContentStyle.Render(ansi.Truncate(line, width, "…")) + "\n")
		}
	}

	return sb.String()
}

// GeneratePreview generates a lightweight preview of the worktree
func (w Worktree) GeneratePreview(width, height int) string {
	var sb strings.Builder

	sb.WriteString(ui.LabelStyle.Render("Path") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Path) + "\n\n")

	sb.WriteString(ui.LabelStyle.Render("Branch") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Branch) + "\n\n")

	sb.WriteString(ui.LabelStyle.Render("Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(w.Head) + "\n")

	if w.IsCurrent {
		sb.WriteString("\n")
		sb.WriteString(ui.LabelStyle.Render("Type") + "\n")
		sb.WriteString(ui.ContentStyle.Render("Current worktree") + "\n")
	}

	return sb.String()
}

// GeneratePreview generates a preview of the commit: metadata plus the
// diffstat from `git show --stat`. The git binary is invoked with an argument
// list (no shell) and the hash comes from git's own log output.
func (c Commit) GeneratePreview(repoPath string, width, height int) string {
	cmd := exec.Command("git", "show", "--stat", "--no-color", "--pretty=format:%an%n%ad%n%n%s%n%n%b", c.Hash)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return ui.ContentStyle.Render(c.Subject)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}

	var sb strings.Builder
	sb.WriteString(ui.LabelStyle.Render("Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(c.Hash) + "\n\n")
	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(ui.ContentStyle.Render(ansi.Truncate(line, width, "…")))
	}
	return sb.String()
}
