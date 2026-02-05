package ports

// NewScanner returns a platform-specific scanner.
// Implemented in scanner_darwin.go, scanner_linux.go, scanner_windows.go.
func NewScanner() Scanner {
	return newPlatformScanner()
}
