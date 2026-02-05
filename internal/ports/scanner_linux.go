//go:build linux

package ports

import "fmt"

type linuxScanner struct{}

func newPlatformScanner() Scanner {
	return &linuxScanner{}
}

func (s *linuxScanner) Scan() ([]PortInfo, error) {
	return nil, fmt.Errorf("linux scanner not yet implemented")
}
