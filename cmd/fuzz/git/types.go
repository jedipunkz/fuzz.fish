package git

// Branch represents a git branch
type Branch struct {
	Name              string
	IsCurrent         bool
	IsRemote          bool
	LastCommit        string
	LastCommitMessage string
	CommitDate        string
	CommitTimestamp   int64 // Unix timestamp for recency scoring
}
