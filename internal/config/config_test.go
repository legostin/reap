package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.RefreshInterval != 2 {
		t.Errorf("expected RefreshInterval=2, got %d", cfg.RefreshInterval)
	}
	if cfg.ShowSystem != false {
		t.Errorf("expected ShowSystem=false, got %v", cfg.ShowSystem)
	}
	if cfg.PortColors == nil {
		t.Error("expected PortColors to be initialized")
	}
	if cfg.PortLabels == nil {
		t.Error("expected PortLabels to be initialized")
	}
	if len(cfg.PortColors) != 0 {
		t.Errorf("expected empty PortColors, got %d entries", len(cfg.PortColors))
	}
	if len(cfg.PortLabels) != 0 {
		t.Errorf("expected empty PortLabels, got %d entries", len(cfg.PortLabels))
	}
}

func TestLoadMissingFile(t *testing.T) {
	// Load should return defaults when config file doesn't exist
	cfg := Load()

	if cfg.RefreshInterval != 2 {
		t.Errorf("expected default RefreshInterval=2, got %d", cfg.RefreshInterval)
	}
	if cfg.ShowSystem != false {
		t.Errorf("expected default ShowSystem=false, got %v", cfg.ShowSystem)
	}
}

func TestLoadValidConfig(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "reap")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `
refresh_interval = 5
show_system = true

[port_colors]
"3000" = "green"
"5432" = "magenta"

[port_labels]
"3000" = "frontend"
"5432" = "database"
`

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Override HOME to use our temp dir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := Load()

	if cfg.RefreshInterval != 5 {
		t.Errorf("expected RefreshInterval=5, got %d", cfg.RefreshInterval)
	}
	if cfg.ShowSystem != true {
		t.Errorf("expected ShowSystem=true, got %v", cfg.ShowSystem)
	}
	if cfg.PortColors["3000"] != "green" {
		t.Errorf("expected PortColors[3000]=green, got %q", cfg.PortColors["3000"])
	}
	if cfg.PortColors["5432"] != "magenta" {
		t.Errorf("expected PortColors[5432]=magenta, got %q", cfg.PortColors["5432"])
	}
	if cfg.PortLabels["3000"] != "frontend" {
		t.Errorf("expected PortLabels[3000]=frontend, got %q", cfg.PortLabels["3000"])
	}
}

func TestLoadInvalidRefreshInterval(t *testing.T) {
	// Create a temp config with invalid refresh interval
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "reap")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `
refresh_interval = 0
show_system = false
`

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := Load()

	// Invalid refresh interval should be corrected to default
	if cfg.RefreshInterval != 2 {
		t.Errorf("expected RefreshInterval=2 for invalid value, got %d", cfg.RefreshInterval)
	}
}

func TestLoadNegativeRefreshInterval(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "reap")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `refresh_interval = -5`

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := Load()

	if cfg.RefreshInterval != 2 {
		t.Errorf("expected RefreshInterval=2 for negative value, got %d", cfg.RefreshInterval)
	}
}

func TestLoadMalformedTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "reap")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `this is not valid toml {{{{`

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Should return defaults on parse error
	cfg := Load()

	if cfg.RefreshInterval != 2 {
		t.Errorf("expected default RefreshInterval=2 on parse error, got %d", cfg.RefreshInterval)
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "reap")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Only set show_system, refresh_interval should use default
	configContent := `show_system = true`

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := Load()

	if cfg.ShowSystem != true {
		t.Errorf("expected ShowSystem=true, got %v", cfg.ShowSystem)
	}
	if cfg.RefreshInterval != 2 {
		t.Errorf("expected default RefreshInterval=2, got %d", cfg.RefreshInterval)
	}
}
