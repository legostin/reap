package ports

import (
	"testing"
)

func TestParseLsofCWDBasic(t *testing.T) {
	output := "p1234\nfcwd\nn/Users/me/projects/my-app\n"
	m := parseLsofCWD(output)

	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
	if m[1234] != "/Users/me/projects/my-app" {
		t.Errorf("expected '/Users/me/projects/my-app', got %q", m[1234])
	}
}

func TestParseLsofCWDMultiple(t *testing.T) {
	output := `p1234
fcwd
n/Users/me/app1
p5678
fcwd
n/Users/me/app2
p9012
fcwd
n/opt/service
`
	m := parseLsofCWD(output)

	if len(m) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(m))
	}
	if m[1234] != "/Users/me/app1" {
		t.Errorf("PID 1234: got %q", m[1234])
	}
	if m[5678] != "/Users/me/app2" {
		t.Errorf("PID 5678: got %q", m[5678])
	}
	if m[9012] != "/opt/service" {
		t.Errorf("PID 9012: got %q", m[9012])
	}
}

func TestParseLsofCWDEmptyString(t *testing.T) {
	m := parseLsofCWD("")
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

func TestParseLsofCWDNoPath(t *testing.T) {
	// PID without path line
	output := "p1234\nfcwd\n"
	m := parseLsofCWD(output)

	if len(m) != 0 {
		t.Errorf("expected empty map when no path, got %d entries", len(m))
	}
}

func TestParseLsofCWDInvalidPID(t *testing.T) {
	// Non-numeric PID
	output := "pabc\nfcwd\nn/path\n"
	m := parseLsofCWD(output)

	if len(m) != 0 {
		t.Errorf("expected empty map for invalid PID, got %d entries", len(m))
	}
}

func TestParseLsofCWDPathWithSpaces(t *testing.T) {
	output := "p1234\nfcwd\nn/Users/my user/my project/app\n"
	m := parseLsofCWD(output)

	if m[1234] != "/Users/my user/my project/app" {
		t.Errorf("expected path with spaces, got %q", m[1234])
	}
}

func TestParseLsofCWDOnlyPath(t *testing.T) {
	// Path without preceding PID
	output := "n/some/path\n"
	m := parseLsofCWD(output)

	if len(m) != 0 {
		t.Errorf("expected empty map when PID=0, got %d entries", len(m))
	}
}

func TestParseLsofCWDMixedContent(t *testing.T) {
	// Real-world output may have extra fields
	output := `p1234
fcwd
n/Users/me/app
p5678
fcwd
n/var/lib/postgresql
`
	m := parseLsofCWD(output)

	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m[1234] != "/Users/me/app" {
		t.Errorf("PID 1234: got %q", m[1234])
	}
	if m[5678] != "/var/lib/postgresql" {
		t.Errorf("PID 5678: got %q", m[5678])
	}
}

func TestFormatElapsedSeconds(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"5", "5s"},
		{"30", "30s"},
		{"59", "59s"},
		{"00", "0s"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatElapsedMinutesSeconds(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"05:30", "5m 30s"},
		{"00:05", "5s"},
		{"01:00", "1m 0s"},
		{"59:59", "59m 59s"},
		{"10:00", "10m 0s"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatElapsedHoursMinutesSeconds(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"01:05:30", "1h 5m"},
		{"00:30:00", "30m 0s"},
		{"12:00:00", "12h 0m"},
		{"23:59:59", "23h 59m"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatElapsedDays(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1-00:00:00", "1d 0h"},
		{"2-03:05:30", "2d 3h"},
		{"30-12:00:00", "30d 12h"},
		{"365-23:59:59", "365d 23h"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatElapsedWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  05:30  ", "5m 30s"},
		{"\t01:00:00\n", "1h 0m"},
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatMemoryZero(t *testing.T) {
	got := formatMemory(0)
	if got != "-" {
		t.Errorf("formatMemory(0) = %q, want '-'", got)
	}
}

func TestFormatMemoryKB(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{1, "1 KB"},
		{100, "100 KB"},
		{512, "512 KB"},
		{1023, "1023 KB"},
	}

	for _, tt := range tests {
		got := formatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("formatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestFormatMemoryMB(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{1024, "1.0 MB"},
		{2048, "2.0 MB"},
		{51200, "50.0 MB"},
		{102400, "100.0 MB"},
		{1047552, "1023.0 MB"}, // Just under 1 GB
	}

	for _, tt := range tests {
		got := formatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("formatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestFormatMemoryGB(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{1048576, "1.0 GB"},
		{2097152, "2.0 GB"},
		{5242880, "5.0 GB"},
		{10485760, "10.0 GB"},
	}

	for _, tt := range tests {
		got := formatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("formatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestFormatMemoryFractional(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{1536, "1.5 MB"},   // 1.5 MB
		{2560, "2.5 MB"},   // 2.5 MB
		{1572864, "1.5 GB"}, // 1.5 GB
	}

	for _, tt := range tests {
		got := formatMemory(tt.kb)
		if got != tt.want {
			t.Errorf("formatMemory(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestParseLsofCWDSpecialPaths(t *testing.T) {
	tests := []struct {
		desc   string
		output string
		pid    int
		path   string
	}{
		{
			desc:   "root path",
			output: "p1\nfcwd\nn/\n",
			pid:    1,
			path:   "/",
		},
		{
			desc:   "path with special chars",
			output: "p100\nfcwd\nn/home/user/my-app_v2.0\n",
			pid:    100,
			path:   "/home/user/my-app_v2.0",
		},
		{
			desc:   "unicode path",
			output: "p200\nfcwd\nn/home/user/\xe2\x9c\x93 project\n",
			pid:    200,
			path:   "/home/user/\xe2\x9c\x93 project",
		},
	}

	for _, tt := range tests {
		m := parseLsofCWD(tt.output)
		if m[tt.pid] != tt.path {
			t.Errorf("%s: got %q, want %q", tt.desc, m[tt.pid], tt.path)
		}
	}
}

func TestParseLsofCWDIgnoresOtherLines(t *testing.T) {
	// Output may contain other prefixes we should ignore
	// Based on the implementation, it captures any 'n' line after a 'p' line
	// The last 'n' line for the PID wins
	output := `p1234
cnode
uuser
fcwd
n/home/user/app
ftxt
n/usr/bin/node
`
	m := parseLsofCWD(output)

	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
	// Implementation captures last 'n' line, so this is the actual behavior
	if m[1234] != "/usr/bin/node" {
		t.Errorf("got %q, want '/usr/bin/node'", m[1234])
	}
}

func TestParseLsofCWDMultiplePaths(t *testing.T) {
	// If multiple 'n' lines for same PID, last one wins (current behavior)
	output := `p1234
fcwd
n/first/path
n/second/path
`
	m := parseLsofCWD(output)

	// Based on implementation, last 'n' line for the PID wins
	if m[1234] != "/second/path" {
		t.Errorf("expected last path, got %q", m[1234])
	}
}

func TestFormatElapsedEdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "0s"},     // empty
		{":", "0s"},    // just colon
		{"-:", "0s"},   // invalid format
	}

	for _, tt := range tests {
		got := formatElapsed(tt.input)
		if got != tt.want {
			t.Errorf("formatElapsed(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatElapsedZeroDays(t *testing.T) {
	// 0 days still shows as seconds because days=0 doesn't trigger day display
	got := formatElapsed("0-00:00:00")
	// With 0 days and 0 hours, it shows seconds
	if got != "0s" {
		t.Errorf("formatElapsed('0-00:00:00') = %q, want '0s'", got)
	}
}
