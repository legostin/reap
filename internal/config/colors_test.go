package config

import (
	"testing"
)

func TestPortColorDefaults(t *testing.T) {
	cfg := Default()

	tests := []struct {
		port  int
		color string
		desc  string
	}{
		// Frontend ports
		{3000, "green", "frontend port 3000"},
		{3001, "green", "frontend port 3001"},
		{4200, "green", "Angular port 4200"},
		{5173, "green", "Vite port 5173"},
		{8080, "green", "frontend port 8080"},
		// Backend ports
		{4000, "yellow", "backend port 4000"},
		{5000, "yellow", "backend port 5000"},
		{8000, "yellow", "backend port 8000"},
		{8081, "yellow", "backend port 8081"},
		{9000, "yellow", "backend port 9000"},
		// Flask / Vite
		{5001, "cyan", "Flask port 5001"},
		{5174, "cyan", "Vite HMR port 5174"},
		{5175, "cyan", "Vite port 5175"},
		// Postgres
		{5432, "magenta", "Postgres port"},
		// Redis
		{6379, "red", "Redis port"},
		// MySQL
		{3306, "blue", "MySQL port"},
		// MongoDB
		{27017, "blue", "MongoDB port"},
		// HTTP/S
		{80, "white", "HTTP port"},
		{443, "white", "HTTPS port"},
		// Unknown port
		{12345, "dim", "unknown port"},
		{1, "dim", "system port"},
	}

	for _, tt := range tests {
		got := cfg.PortColor(tt.port)
		if got != tt.color {
			t.Errorf("PortColor(%d) = %q, want %q (%s)", tt.port, got, tt.color, tt.desc)
		}
	}
}

func TestPortColorUserOverride(t *testing.T) {
	cfg := Default()
	cfg.PortColors = map[string]string{
		"3000": "blue",   // Override default green
		"9999": "purple", // Custom port
	}

	tests := []struct {
		port  int
		color string
		desc  string
	}{
		{3000, "blue", "user override for 3000"},
		{9999, "purple", "custom port color"},
		{5432, "magenta", "default still works"},
		{12345, "dim", "unknown port"},
	}

	for _, tt := range tests {
		got := cfg.PortColor(tt.port)
		if got != tt.color {
			t.Errorf("PortColor(%d) = %q, want %q (%s)", tt.port, got, tt.color, tt.desc)
		}
	}
}

func TestPortColorUserOverridePriority(t *testing.T) {
	cfg := Default()

	// User override should take precedence over default
	cfg.PortColors = map[string]string{
		"80": "red", // Override HTTP from white to red
	}

	got := cfg.PortColor(80)
	if got != "red" {
		t.Errorf("user override should take precedence: got %q, want red", got)
	}
}

func TestPortColorEmptyUserOverrides(t *testing.T) {
	cfg := Default()
	cfg.PortColors = map[string]string{}

	// All defaults should still work
	if got := cfg.PortColor(3000); got != "green" {
		t.Errorf("PortColor(3000) = %q, want green", got)
	}
	if got := cfg.PortColor(5432); got != "magenta" {
		t.Errorf("PortColor(5432) = %q, want magenta", got)
	}
}

func TestPortColorNilUserOverrides(t *testing.T) {
	cfg := Default()
	cfg.PortColors = nil

	// Should not panic and return defaults
	if got := cfg.PortColor(3000); got != "green" {
		t.Errorf("PortColor(3000) = %q, want green", got)
	}
}

func TestPortKey(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{0, "0"},
		{3000, "3000"},
		{65535, "65535"},
	}

	for _, tt := range tests {
		got := portKey(tt.port)
		if got != tt.want {
			t.Errorf("portKey(%d) = %q, want %q", tt.port, got, tt.want)
		}
	}
}

func TestDefaultPortColorsCompleteness(t *testing.T) {
	// Ensure all expected default ports are defined
	expectedPorts := []int{
		3000, 3001, 4200, 5173, 8080, // frontend
		4000, 5000, 8000, 8081, 9000, // backend
		5001, 5174, 5175, // flask/vite
		5432,  // postgres
		6379,  // redis
		3306,  // mysql
		27017, // mongodb
		80, 443, // http/s
	}

	for _, port := range expectedPorts {
		if _, ok := defaultPortColors[port]; !ok {
			t.Errorf("expected port %d to be in defaultPortColors", port)
		}
	}
}

func TestAllColorCategories(t *testing.T) {
	// Test that all color categories are represented
	colors := make(map[string]bool)
	for _, color := range defaultPortColors {
		colors[color] = true
	}

	expected := []string{"green", "yellow", "cyan", "magenta", "red", "blue", "white"}
	for _, c := range expected {
		if !colors[c] {
			t.Errorf("expected color %q to be used in defaultPortColors", c)
		}
	}
}
