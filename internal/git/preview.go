package git

import (
	"os/exec"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// branchRecentCommits is how many `git log --oneline` style commits the
// branch and worktree previews list.
const branchRecentCommits = 5

// writeRecentCommits appends the Recent commits section: the hash and the
// subject are color-coded with subdued theme colors, and lines fold instead
// of truncating (summaries may hold long or multibyte subjects with no
// spaces for word wrapping; Hardwrap keeps the ANSI coloring intact).
// Branches whose commit cannot be read render no section.
func writeRecentCommits(sb *strings.Builder, repoPath, hash string, width int) {
	repo := Repository{Path: repoPath}
	commits := repo.RecentCommits(hash, branchRecentCommits)
	if len(commits) == 0 {
		return
	}

	sb.WriteString("\n")
	sb.WriteString(ui.ContextHeaderStyle.Render("Recent commits") + "\n")
	for _, c := range commits {
		line := ui.CommitHashStyle.Render(c.Hash) + " " +
			ui.CommitSubjectStyle.Render(strings.ReplaceAll(c.Subject, "\n", " "))
		sb.WriteString(ansi.Hardwrap(line, width, false))
		sb.WriteString("\n")
	}
}

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

	writeRecentCommits(&sb, repoPath, b.Hash, width)

	return sb.String()
}

// GeneratePreview generates a preview of the worktree: metadata plus the
// worktree's last commits. The walk anchors on the worktree HEAD's full
// commit hash; a bare or unborn worktree keeps the metadata-only preview.
func (w Worktree) GeneratePreview(repoPath string, width, height int) string {
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

	writeRecentCommits(&sb, repoPath, w.Hash, width)

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
