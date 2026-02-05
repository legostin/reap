package ports

// PortInfo represents a listening port and its associated process.
type PortInfo struct {
	Port      int
	PID       int
	PPID      int    // parent process ID
	Process   string
	User      string
	Command   string
	Protocol  string
	Address   string
	Uptime    string
	Memory    string // human-readable, e.g. "12.3 MB"
	Container string // Docker container name, empty if not in Docker
	CWD       string // working directory of the process
}

// Scanner discovers listening ports on the system.
type Scanner interface {
	Scan() ([]PortInfo, error)
}
