package ports

import (
	"os/exec"
	"strconv"
	"strings"
)

// dockerPortMap maps host port -> container name by parsing `docker ps`.
type dockerPortMap map[int]string

// detectDocker runs `docker ps --format` and builds a map of
// host port -> container name. Returns nil if docker is unavailable.
func detectDocker() dockerPortMap {
	out, err := exec.Command("docker", "ps",
		"--format", "{{.Names}}\t{{.Ports}}").Output()
	if err != nil {
		return nil
	}
	return parseDockerPS(string(out))
}

// parseDockerPS parses `docker ps --format "{{.Names}}\t{{.Ports}}"` output.
// Ports look like: "0.0.0.0:3000->3000/tcp, :::3000->3000/tcp"
func parseDockerPS(output string) dockerPortMap {
	m := make(dockerPortMap)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		portMappings := parts[1]

		for _, mapping := range strings.Split(portMappings, ", ") {
			// "0.0.0.0:3000->3000/tcp" or ":::3000->3000/tcp"
			hostPort := extractHostPort(mapping)
			if hostPort > 0 {
				m[hostPort] = name
			}
		}
	}
	return m
}

// extractHostPort extracts the host port from a mapping like "0.0.0.0:3000->3000/tcp".
func extractHostPort(mapping string) int {
	arrow := strings.Index(mapping, "->")
	if arrow == -1 {
		return 0
	}
	hostPart := mapping[:arrow] // "0.0.0.0:3000" or ":::3000"

	lastColon := strings.LastIndex(hostPart, ":")
	if lastColon == -1 {
		return 0
	}
	port, err := strconv.Atoi(hostPart[lastColon+1:])
	if err != nil {
		return 0
	}
	return port
}

// enrichDockerInfo fills Container field for ports that match Docker port mappings.
func enrichDockerInfo(ports []PortInfo) {
	dm := detectDocker()
	if dm == nil {
		return
	}
	for i := range ports {
		if name, ok := dm[ports[i].Port]; ok {
			ports[i].Container = name
		}
	}
}
