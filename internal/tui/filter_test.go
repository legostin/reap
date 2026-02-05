package tui

import (
	"testing"

	"github.com/legostin/reap/internal/ports"
)

func TestNewFilterInput(t *testing.T) {
	f := newFilterInput()

	if f.active {
		t.Error("new filter should not be active")
	}
	if f.value() != "" {
		t.Errorf("new filter should have empty value, got %q", f.value())
	}
}

func TestFilterActivate(t *testing.T) {
	f := newFilterInput()
	f.activate()

	if !f.active {
		t.Error("filter should be active after activate()")
	}
}

func TestFilterDeactivate(t *testing.T) {
	f := newFilterInput()
	f.activate()
	f.deactivate()

	if f.active {
		t.Error("filter should not be active after deactivate()")
	}
}

func TestFilterClear(t *testing.T) {
	f := newFilterInput()
	f.activate()
	f.input.SetValue("test query")
	f.clear()

	if f.active {
		t.Error("filter should not be active after clear()")
	}
	if f.value() != "" {
		t.Errorf("filter value should be empty after clear(), got %q", f.value())
	}
}

func TestFilterValue(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("node")

	if f.value() != "node" {
		t.Errorf("expected value 'node', got %q", f.value())
	}
}

func TestFilterMatchesEmptyQuery(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("")

	p := ports.PortInfo{
		Port:    3000,
		PID:     1234,
		Process: "node",
		User:    "user",
	}

	if !f.matches(p) {
		t.Error("empty query should match everything")
	}
}

func TestFilterMatchesProcess(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("node")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{Process: "node"},
			match: true,
			desc:  "exact match",
		},
		{
			port:  ports.PortInfo{Process: "nodejs"},
			match: true,
			desc:  "partial match",
		},
		{
			port:  ports.PortInfo{Process: "Node"},
			match: true,
			desc:  "case insensitive",
		},
		{
			port:  ports.PortInfo{Process: "NODE"},
			match: true,
			desc:  "uppercase",
		},
		{
			port:  ports.PortInfo{Process: "python"},
			match: false,
			desc:  "no match",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(%q) = %v, want %v", tt.desc, tt.port.Process, got, tt.match)
		}
	}
}

func TestFilterMatchesPort(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("3000")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{Port: 3000},
			match: true,
			desc:  "exact port match",
		},
		{
			port:  ports.PortInfo{Port: 30000},
			match: true,
			desc:  "partial port match (contains 3000)",
		},
		{
			port:  ports.PortInfo{Port: 8080},
			match: false,
			desc:  "no match",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(port=%d) = %v, want %v", tt.desc, tt.port.Port, got, tt.match)
		}
	}
}

func TestFilterMatchesPID(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("1234")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{PID: 1234},
			match: true,
			desc:  "exact PID match",
		},
		{
			port:  ports.PortInfo{PID: 12345},
			match: true,
			desc:  "partial PID match",
		},
		{
			port:  ports.PortInfo{PID: 5678},
			match: false,
			desc:  "no match",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(PID=%d) = %v, want %v", tt.desc, tt.port.PID, got, tt.match)
		}
	}
}

func TestFilterMatchesUser(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("root")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{User: "root"},
			match: true,
			desc:  "exact user match",
		},
		{
			port:  ports.PortInfo{User: "ROOT"},
			match: true,
			desc:  "case insensitive",
		},
		{
			port:  ports.PortInfo{User: "user"},
			match: false,
			desc:  "no match",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(User=%q) = %v, want %v", tt.desc, tt.port.User, got, tt.match)
		}
	}
}

func TestFilterMatchesContainer(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("redis")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{Container: "my-redis"},
			match: true,
			desc:  "container match",
		},
		{
			port:  ports.PortInfo{Container: "Redis-Server"},
			match: true,
			desc:  "case insensitive",
		},
		{
			port:  ports.PortInfo{Container: "postgres"},
			match: false,
			desc:  "no match",
		},
		{
			port:  ports.PortInfo{Container: ""},
			match: false,
			desc:  "empty container",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(Container=%q) = %v, want %v", tt.desc, tt.port.Container, got, tt.match)
		}
	}
}

func TestFilterMatchesCWD(t *testing.T) {
	f := newFilterInput()
	f.input.SetValue("myapp")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{CWD: "/home/user/myapp"},
			match: true,
			desc:  "CWD match",
		},
		{
			port:  ports.PortInfo{CWD: "/Users/dev/MyApp/src"},
			match: true,
			desc:  "case insensitive",
		},
		{
			port:  ports.PortInfo{CWD: "/opt/other"},
			match: false,
			desc:  "no match",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: matches(CWD=%q) = %v, want %v", tt.desc, tt.port.CWD, got, tt.match)
		}
	}
}

func TestFilterMatchesMultipleFields(t *testing.T) {
	// Filter matches if ANY field contains the query
	f := newFilterInput()
	f.input.SetValue("test")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{Process: "test-server"},
			match: true,
			desc:  "match in process",
		},
		{
			port:  ports.PortInfo{User: "tester"},
			match: true,
			desc:  "match in user",
		},
		{
			port:  ports.PortInfo{Container: "test-container"},
			match: true,
			desc:  "match in container",
		},
		{
			port:  ports.PortInfo{CWD: "/path/to/test/dir"},
			match: true,
			desc:  "match in CWD",
		},
		{
			port:  ports.PortInfo{Process: "node", User: "user", Container: "", CWD: "/app"},
			match: false,
			desc:  "no match in any field",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.match)
		}
	}
}

func TestFilterMatchesCaseInsensitiveQuery(t *testing.T) {
	// Query is also lowercased before comparison
	tests := []struct {
		query   string
		process string
		match   bool
	}{
		{"NODE", "node", true},
		{"Node", "NODE", true},
		{"nOdE", "NoDE", true},
	}

	for _, tt := range tests {
		f := newFilterInput()
		f.input.SetValue(tt.query)

		p := ports.PortInfo{Process: tt.process}
		got := f.matches(p)
		if got != tt.match {
			t.Errorf("query=%q process=%q: got %v, want %v", tt.query, tt.process, got, tt.match)
		}
	}
}

func TestFilterMatchesNumericInText(t *testing.T) {
	// Numbers in query can match process names containing numbers
	f := newFilterInput()
	f.input.SetValue("3")

	tests := []struct {
		port  ports.PortInfo
		match bool
		desc  string
	}{
		{
			port:  ports.PortInfo{Port: 3000},
			match: true,
			desc:  "port starts with 3",
		},
		{
			port:  ports.PortInfo{Process: "python3"},
			match: true,
			desc:  "process contains 3",
		},
		{
			port:  ports.PortInfo{PID: 123},
			match: true,
			desc:  "PID contains 3",
		},
	}

	for _, tt := range tests {
		got := f.matches(tt.port)
		if got != tt.match {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.match)
		}
	}
}
