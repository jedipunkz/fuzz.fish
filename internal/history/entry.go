package history

// Entry represents a single command from Fish shell history
type Entry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
	// Count is how many times the command appears in the history file. Entries
	// are deduplicated during parsing, so this is the only place the raw
	// frequency survives.
	Count int
}
