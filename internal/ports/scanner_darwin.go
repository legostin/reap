//go:build darwin

package ports

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type darwinScanner struct{}

func newPlatformScanner() Scanner {
	return &darwinScanner{}
}

func (s *darwinScanner) Scan() ([]PortInfo, error) {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	ports := parseLsofOutput(string(out))
	if len(ports) > 0 {
		enrichProcessInfo(ports)
		enrichDockerInfo(ports)
	}
	return ports, nil
}

// parseLsofOutput parses lsof tabular output into PortInfo entries.
// Deduplicates by (port, PID), preferring IPv4 over IPv6.
func parseLsofOutput(output string) []PortInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	type dedupKey struct {
		port int
		pid  int
	}
	seen := make(map[dedupKey]int) // key -> index in result
	var result []PortInfo

	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		process := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		user := fields[2]
		protocol := strings.ToLower(fields[7]) // TCP -> tcp

		// NAME field is after NODE (index 8). It may be followed by "(LISTEN)".
		name := fields[8]
		addr, portStr := parseNameField(name)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		key := dedupKey{port: port, pid: pid}
		isIPv4 := !strings.Contains(addr, ":")

		if idx, exists := seen[key]; exists {
			// prefer IPv4 over IPv6
			if isIPv4 && strings.Contains(result[idx].Address, ":") {
				result[idx].Address = addr
			}
			continue
		}

		seen[key] = len(result)
		result = append(result, PortInfo{
			Port:     port,
			PID:      pid,
			Process:  process,
			User:     user,
			Protocol: protocol,
			Address:  addr,
		})
	}

	return result
}

// parseNameField splits "addr:port" from lsof NAME column.
// Handles IPv6 like "[::1]:8080" and IPv4 like "127.0.0.1:8080" or "*:8080".
func parseNameField(name string) (addr, port string) {
	// Remove any trailing state info like "(LISTEN)"
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}

	if strings.HasPrefix(name, "[") {
		// IPv6: [::1]:8080
		if closeBracket := strings.LastIndex(name, "]"); closeBracket != -1 {
			addr = name[:closeBracket+1]
			if closeBracket+2 < len(name) {
				port = name[closeBracket+2:] // skip ]:
			}
			return addr, port
		}
	}

	// IPv4 or *: last colon separates addr:port
	if lastColon := strings.LastIndex(name, ":"); lastColon != -1 {
		return name[:lastColon], name[lastColon+1:]
	}
	return name, ""
}
