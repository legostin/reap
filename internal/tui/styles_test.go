package tui

import (
	"strings"
	"testing"
)

func TestPortStyle(t *testing.T) {
	// Test all valid color names
	validColors := []string{"green", "yellow", "cyan", "magenta", "red", "blue", "white", "dim"}

	for _, color := range validColors {
		style := portStyle(color)
		// Style should be non-zero (have properties set)
		rendered := style.Render("test")
		if rendered == "" {
			t.Errorf("portStyle(%q) should render text", color)
		}
	}
}

func TestPortStyleUnknownColor(t *testing.T) {
	// Unknown color should fall back to "dim"
	style := portStyle("nonexistent")
	dimStyle := portStyle("dim")

	// Both should render the same way (fallback to dim)
	rendered := style.Render("test")
	dimRendered := dimStyle.Render("test")

	if rendered != dimRendered {
		t.Errorf("unknown color should fall back to dim")
	}
}

func TestPortStyleEmpty(t *testing.T) {
	// Empty color name should fall back to dim
	style := portStyle("")
	rendered := style.Render("test")
	if rendered == "" {
		t.Error("empty color should still render text")
	}
}

func TestColorLegend(t *testing.T) {
	legend := colorLegend()

	// Should contain all category labels
	expectedLabels := []string{
		"frontend",
		"backend",
		"flask/vite",
		"postgres",
		"redis",
		"mysql/mongo",
		"http/s",
		"other",
	}

	for _, label := range expectedLabels {
		if !strings.Contains(legend, label) {
			t.Errorf("colorLegend() should contain %q", label)
		}
	}
}

func TestColorLegendFormat(t *testing.T) {
	legend := colorLegend()

	// Should contain dot characters (●)
	if !strings.Contains(legend, "●") {
		t.Error("colorLegend() should contain dot characters")
	}

	// Should have multiple entries separated by spaces
	if len(legend) < 50 {
		t.Errorf("colorLegend() seems too short: %d chars", len(legend))
	}
}

func TestPortColorMapCompleteness(t *testing.T) {
	// Verify all expected colors exist in the map
	expectedColors := []string{"green", "yellow", "cyan", "magenta", "red", "blue", "white", "dim"}

	for _, color := range expectedColors {
		if _, ok := portColorMap[color]; !ok {
			t.Errorf("portColorMap missing color %q", color)
		}
	}
}

func TestStylesNonNil(t *testing.T) {
	// Test that all styles are usable and don't panic
	styles := []struct {
		name  string
		style func() string
	}{
		{"headerStyle", func() string { return headerStyle.Render("test") }},
		{"statusBarStyle", func() string { return statusBarStyle.Render("test") }},
		{"selectedRowStyle", func() string { return selectedRowStyle.Render("test") }},
		{"dialogStyle", func() string { return dialogStyle.Render("test") }},
		{"dialogTitleStyle", func() string { return dialogTitleStyle.Render("test") }},
		{"errorStyle", func() string { return errorStyle.Render("test") }},
		{"successStyle", func() string { return successStyle.Render("test") }},
		{"filterPromptStyle", func() string { return filterPromptStyle.Render("test") }},
		{"tableHeaderStyle", func() string { return tableHeaderStyle.Render("test") }},
		{"cellStyle", func() string { return cellStyle.Render("test") }},
		{"childCellStyle", func() string { return childCellStyle.Render("test") }},
		{"expandLabelStyle", func() string { return expandLabelStyle.Render("test") }},
		{"expandValueStyle", func() string { return expandValueStyle.Render("test") }},
		{"parentStyle", func() string { return parentStyle.Render("test") }},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			result := s.style()
			if result == "" {
				t.Errorf("%s.Render() returned empty string", s.name)
			}
		})
	}
}
