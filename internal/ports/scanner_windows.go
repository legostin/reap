//go:build windows

package ports

import "fmt"

type windowsScanner struct{}

func newPlatformScanner() Scanner {
	return &windowsScanner{}
}

func (s *windowsScanner) Scan() ([]PortInfo, error) {
	return nil, fmt.Errorf("windows scanner not yet implemented")
}
