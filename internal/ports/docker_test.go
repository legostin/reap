package ports

import (
	"testing"
)

func TestParseDockerPSBasic(t *testing.T) {
	output := "my-redis\t0.0.0.0:6379->6379/tcp"
	m := parseDockerPS(output)

	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
	if m[6379] != "my-redis" {
		t.Errorf("expected my-redis, got %q", m[6379])
	}
}

func TestParseDockerPSMultiplePorts(t *testing.T) {
	output := "web-app\t0.0.0.0:3000->3000/tcp, 0.0.0.0:3001->3001/tcp"
	m := parseDockerPS(output)

	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m[3000] != "web-app" {
		t.Errorf("port 3000: expected web-app, got %q", m[3000])
	}
	if m[3001] != "web-app" {
		t.Errorf("port 3001: expected web-app, got %q", m[3001])
	}
}

func TestParseDockerPSIPv6(t *testing.T) {
	output := "my-service\t:::8080->8080/tcp"
	m := parseDockerPS(output)

	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
	if m[8080] != "my-service" {
		t.Errorf("expected my-service, got %q", m[8080])
	}
}

func TestParseDockerPSMixedIPv4IPv6(t *testing.T) {
	output := "my-redis\t0.0.0.0:6379->6379/tcp, :::6379->6379/tcp"
	m := parseDockerPS(output)

	// Same port mapped for both IPv4 and IPv6, should still be one entry
	if m[6379] != "my-redis" {
		t.Errorf("expected my-redis, got %q", m[6379])
	}
}

func TestParseDockerPSMultipleContainers(t *testing.T) {
	output := `my-redis	0.0.0.0:6379->6379/tcp
my-postgres	0.0.0.0:5432->5432/tcp
web-app	0.0.0.0:3000->3000/tcp, 0.0.0.0:3001->3001/tcp`

	m := parseDockerPS(output)

	if len(m) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(m))
	}
	if m[6379] != "my-redis" {
		t.Errorf("port 6379: got %q", m[6379])
	}
	if m[5432] != "my-postgres" {
		t.Errorf("port 5432: got %q", m[5432])
	}
	if m[3000] != "web-app" {
		t.Errorf("port 3000: got %q", m[3000])
	}
	if m[3001] != "web-app" {
		t.Errorf("port 3001: got %q", m[3001])
	}
}

func TestParseDockerPSEmptyOutput(t *testing.T) {
	m := parseDockerPS("")
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestParseDockerPSWhitespaceOnly(t *testing.T) {
	m := parseDockerPS("   \n\t\n   ")
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestParseDockerPSNoPortMappings(t *testing.T) {
	output := "my-container\t"
	m := parseDockerPS(output)
	if len(m) != 0 {
		t.Errorf("expected empty map for container with no ports, got %d", len(m))
	}
}

func TestParseDockerPSMalformedLine(t *testing.T) {
	// Line without tab separator
	output := "malformed-no-tab"
	m := parseDockerPS(output)
	if len(m) != 0 {
		t.Errorf("expected empty map for malformed line, got %d", len(m))
	}
}

func TestExtractHostPortBasic(t *testing.T) {
	tests := []struct {
		mapping string
		want    int
	}{
		{"0.0.0.0:3000->3000/tcp", 3000},
		{"127.0.0.1:8080->80/tcp", 8080},
		{":::6379->6379/tcp", 6379},
		{"[::]:5432->5432/tcp", 5432},
	}

	for _, tt := range tests {
		got := extractHostPort(tt.mapping)
		if got != tt.want {
			t.Errorf("extractHostPort(%q) = %d, want %d", tt.mapping, got, tt.want)
		}
	}
}

func TestExtractHostPortInvalid(t *testing.T) {
	tests := []struct {
		mapping string
		desc    string
	}{
		{"3000/tcp", "no arrow"},
		{"", "empty"},
		{"->3000/tcp", "no host part"},
		{"abc->3000/tcp", "no colon in host"},
		{"0.0.0.0:abc->3000/tcp", "non-numeric port"},
	}

	for _, tt := range tests {
		got := extractHostPort(tt.mapping)
		if got != 0 {
			t.Errorf("extractHostPort(%q) = %d, want 0 (%s)", tt.mapping, got, tt.desc)
		}
	}
}

func TestExtractHostPortDifferentHostContainerPorts(t *testing.T) {
	// Host port can differ from container port
	got := extractHostPort("0.0.0.0:8080->80/tcp")
	if got != 8080 {
		t.Errorf("expected host port 8080, got %d", got)
	}
}

func TestExtractHostPortHighPort(t *testing.T) {
	got := extractHostPort("0.0.0.0:65535->65535/tcp")
	if got != 65535 {
		t.Errorf("expected port 65535, got %d", got)
	}
}

func TestParseDockerPSInternalPorts(t *testing.T) {
	// Container with internal-only port (no host mapping)
	output := "my-container\t3000/tcp"
	m := parseDockerPS(output)

	// Internal ports without -> should not be included
	if len(m) != 0 {
		t.Errorf("internal ports should not be mapped, got %d entries", len(m))
	}
}

func TestDockerPortMapType(t *testing.T) {
	// Test that dockerPortMap is usable
	var m dockerPortMap = make(map[int]string)
	m[3000] = "test-container"

	if m[3000] != "test-container" {
		t.Error("dockerPortMap should work as map[int]string")
	}

	// Test non-existent port
	if name, ok := m[9999]; ok {
		t.Errorf("non-existent port should return false, got %q", name)
	}
}

func TestParseDockerPSContainerNamesWithSpecialChars(t *testing.T) {
	tests := []struct {
		output string
		name   string
		port   int
	}{
		{"my-app_web_1\t0.0.0.0:3000->3000/tcp", "my-app_web_1", 3000},
		{"project.service\t0.0.0.0:8080->8080/tcp", "project.service", 8080},
		{"container-with-dashes\t0.0.0.0:5000->5000/tcp", "container-with-dashes", 5000},
	}

	for _, tt := range tests {
		m := parseDockerPS(tt.output)
		if m[tt.port] != tt.name {
			t.Errorf("expected container name %q for port %d, got %q", tt.name, tt.port, m[tt.port])
		}
	}
}

func TestParseDockerPSUDP(t *testing.T) {
	// UDP port mappings
	output := "my-dns\t0.0.0.0:53->53/udp"
	m := parseDockerPS(output)

	if m[53] != "my-dns" {
		t.Errorf("expected my-dns for UDP port, got %q", m[53])
	}
}

func TestParseDockerPSMixedProtocols(t *testing.T) {
	output := "my-service\t0.0.0.0:53->53/tcp, 0.0.0.0:53->53/udp"
	m := parseDockerPS(output)

	// Same port with different protocols - last one wins
	if m[53] == "" {
		t.Error("expected port 53 to be mapped")
	}
}
