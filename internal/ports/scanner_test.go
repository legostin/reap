package ports

import (
	"testing"
)

const mockLsofOutput = `COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node      1234   user   23u  IPv4 0x1234567890      0t0  TCP *:3000 (LISTEN)
node      1234   user   24u  IPv6 0x1234567891      0t0  TCP [::1]:3000 (LISTEN)
postgres  5678   user   10u  IPv4 0x1234567892      0t0  TCP 127.0.0.1:5432 (LISTEN)
python3   9012   root   5u   IPv4 0x1234567893      0t0  TCP *:8000 (LISTEN)
ruby      3456   user   8u   IPv6 0x1234567894      0t0  TCP [::]:3001 (LISTEN)
`

func TestParseLsofOutput(t *testing.T) {
	ports := parseLsofOutput(mockLsofOutput)

	if len(ports) != 4 {
		t.Fatalf("expected 4 ports, got %d", len(ports))
	}

	// Check dedup: node on port 3000 should appear once, prefer IPv4
	found := false
	for _, p := range ports {
		if p.Port == 3000 && p.PID == 1234 {
			found = true
			if p.Address != "*" {
				t.Errorf("expected IPv4 address '*', got %q", p.Address)
			}
			if p.Process != "node" {
				t.Errorf("expected process 'node', got %q", p.Process)
			}
		}
	}
	if !found {
		t.Error("node on port 3000 not found")
	}

	// Check postgres
	for _, p := range ports {
		if p.Port == 5432 {
			if p.PID != 5678 {
				t.Errorf("expected PID 5678, got %d", p.PID)
			}
			if p.Address != "127.0.0.1" {
				t.Errorf("expected address '127.0.0.1', got %q", p.Address)
			}
			if p.User != "user" {
				t.Errorf("expected user 'user', got %q", p.User)
			}
		}
	}
}

func TestParseLsofOutputEmpty(t *testing.T) {
	ports := parseLsofOutput("")
	if len(ports) != 0 {
		t.Fatalf("expected 0 ports, got %d", len(ports))
	}

	ports = parseLsofOutput("COMMAND     PID   USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME\n")
	if len(ports) != 0 {
		t.Fatalf("expected 0 ports for header-only, got %d", len(ports))
	}
}

func TestParseNameField(t *testing.T) {
	tests := []struct {
		input    string
		wantAddr string
		wantPort string
	}{
		{"*:3000", "*", "3000"},
		{"127.0.0.1:5432", "127.0.0.1", "5432"},
		{"[::1]:8080", "[::1]", "8080"},
		{"[::]:3001", "[::]", "3001"},
	}

	for _, tt := range tests {
		addr, port := parseNameField(tt.input)
		if addr != tt.wantAddr || port != tt.wantPort {
			t.Errorf("parseNameField(%q) = (%q, %q), want (%q, %q)",
				tt.input, addr, port, tt.wantAddr, tt.wantPort)
		}
	}
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"05:30", "5m 30s"},
		{"01:05:30", "1h 5m"},
		{"2-03:05:30", "2d 3h"},
		{"30", "30s"},
		{"00:05", "5s"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseDockerPS(t *testing.T) {
	output := `my-redis	0.0.0.0:6379->6379/tcp, :::6379->6379/tcp
my-postgres	0.0.0.0:5432->5432/tcp
web-app	0.0.0.0:3000->3000/tcp, 0.0.0.0:3001->3001/tcp`

	m := parseDockerPS(output)

	tests := []struct {
		port int
		want string
	}{
		{6379, "my-redis"},
		{5432, "my-postgres"},
		{3000, "web-app"},
		{3001, "web-app"},
	}

	for _, tt := range tests {
		got, ok := m[tt.port]
		if !ok {
			t.Errorf("port %d not found in docker map", tt.port)
			continue
		}
		if got != tt.want {
			t.Errorf("port %d: got %q, want %q", tt.port, got, tt.want)
		}
	}

	if _, ok := m[9999]; ok {
		t.Error("unexpected port 9999 in docker map")
	}
}

func TestParseDockerPSEmpty(t *testing.T) {
	m := parseDockerPS("")
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0.0.0.0:3000->3000/tcp", 3000},
		{":::6379->6379/tcp", 6379},
		{"127.0.0.1:5432->5432/tcp", 5432},
		{"3000/tcp", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := extractHostPort(tt.input)
		if got != tt.want {
			t.Errorf("extractHostPort(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseLsofCWD(t *testing.T) {
	output := "p1234\nfcwd\nn/Users/me/projects/my-app\np5678\nfcwd\nn/opt/homebrew/var\n"
	m := parseLsofCWD(output)

	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m[1234] != "/Users/me/projects/my-app" {
		t.Errorf("pid 1234: got %q", m[1234])
	}
	if m[5678] != "/opt/homebrew/var" {
		t.Errorf("pid 5678: got %q", m[5678])
	}
}

func TestParseLsofCWDEmpty(t *testing.T) {
	m := parseLsofCWD("")
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestFormatMemory(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{0, "-"},
		{512, "512 KB"},
		{1024, "1.0 MB"},
		{51200, "50.0 MB"},
		{1048576, "1.0 GB"},
	}

	for _, tt := range tests {
		got := formatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("formatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}
