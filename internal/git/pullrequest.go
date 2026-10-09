package git

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// pullRequestLimit caps how many open pull requests are requested from gh.
// ponytail: fixed cap, page through `gh pr list` if repos exceed it.
const pullRequestLimit = 200

// PullRequest is an open pull request of the current repository.
type PullRequest struct {
	Number     int
	Title      string
	URL        string
	Repository string // base repository as owner/name
	Author     string
	Branch     string // head branch name; "owner:branch" for fork PRs
	Worktree   string // local worktree with the head branch checked out, "" if none
}

type ghPullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	HeadRefName         string `json:"headRefName"`
	HeadRepositoryOwner struct {
		Login string `json:"login"`
	} `json:"headRepositoryOwner"`
	IsCrossRepository bool `json:"isCrossRepository"`
}

// PullRequests lists the open pull requests of the repository with
// `gh pr list`. gh must be installed and authenticated (`gh auth login`).
// Commands run with argument lists (no shell).
func (r *Repository) PullRequests() ([]PullRequest, error) {
	cmd := exec.Command("gh", "pr", "list", "--state", "open",
		"--limit", strconv.Itoa(pullRequestLimit),
		"--json", "number,title,url,author,headRefName,headRepositoryOwner,isCrossRepository")
	cmd.Dir = r.Path
	out, err := cmd.Output()
	if errors.Is(err, exec.ErrNotFound) {
		return nil, errors.New("gh command not found")
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		// gh explains the failure (e.g. "run gh auth login") on stderr.
		msg, _, _ := strings.Cut(strings.TrimSpace(string(exitErr.Stderr)), "\n")
		if msg != "" {
			return nil, errors.New("gh: " + msg)
		}
	}
	if err != nil {
		return nil, err
	}

	worktrees, _ := r.Worktrees()
	return parsePullRequests(out, worktrees)
}

// parsePullRequests decodes `gh pr list --json` output and links each pull
// request to the worktree that has its head branch checked out.
func parsePullRequests(data []byte, worktrees []Worktree) ([]PullRequest, error) {
	var raw []ghPullRequest
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	branchPath := make(map[string]string, len(worktrees))
	for _, w := range worktrees {
		branchPath[w.Branch] = w.Path
	}

	prs := make([]PullRequest, 0, len(raw))
	for _, p := range raw {
		pr := PullRequest{
			Number:     p.Number,
			Title:      p.Title,
			URL:        p.URL,
			Repository: repositoryFromURL(p.URL),
			Author:     p.Author.Login,
			Branch:     p.HeadRefName,
		}
		if p.IsCrossRepository {
			// A fork's branch name says nothing about local branches: a fork
			// PR from "main" must not match the local main worktree.
			pr.Branch = p.HeadRepositoryOwner.Login + ":" + p.HeadRefName
		} else {
			pr.Worktree = branchPath[p.HeadRefName]
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// repositoryFromURL extracts "owner/name" from a pull request URL such as
// https://github.com/owner/name/pull/1.
func repositoryFromURL(url string) string {
	_, path, ok := strings.Cut(url, "://")
	if !ok {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[1] + "/" + parts[2]
}

// GeneratePreview renders the pull request details. Pure string building.
func (pr PullRequest) GeneratePreview(width int) string {
	worktree := pr.Worktree
	if worktree == "" {
		worktree = "-"
	}

	var sb strings.Builder
	for i, f := range []struct{ label, value string }{
		{"Pull Request", "#" + strconv.Itoa(pr.Number) + " " + pr.Title + "\n" + pr.URL},
		{"Repository", pr.Repository},
		{"Author", pr.Author},
		{"Branch", pr.Branch},
		{"Worktree", worktree},
	} {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(ui.LabelStyle.Render(f.label))
		for _, line := range strings.Split(f.value, "\n") {
			sb.WriteString("\n" + ui.ContentStyle.Render(ansi.Truncate(line, width, "…")))
		}
	}
	return sb.String()
}
