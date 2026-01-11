package history

// Entry represents a single command from Fish shell history
type Entry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
}
