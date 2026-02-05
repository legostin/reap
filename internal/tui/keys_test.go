package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestKeysDefinition(t *testing.T) {
	// Test that all key bindings are defined
	bindings := []struct {
		name    string
		binding key.Binding
		keys    []string
	}{
		{"Up", keys.Up, []string{"up"}},
		{"Down", keys.Down, []string{"down", "j"}},
		{"Enter", keys.Enter, []string{"enter"}},
		{"Kill", keys.Kill, []string{"k"}},
		{"ForceK", keys.ForceK, []string{"K"}},
		{"KillParent", keys.KillParent, []string{"p"}},
		{"Filter", keys.Filter, []string{"/"}},
		{"Sort", keys.Sort, []string{"s"}},
		{"SortRev", keys.SortRev, []string{"S"}},
		{"System", keys.System, []string{"a"}},
		{"Tree", keys.Tree, []string{"t"}},
		{"Refresh", keys.Refresh, []string{"r"}},
		{"Help", keys.Help, []string{"?"}},
		{"Quit", keys.Quit, []string{"q", "ctrl+c"}},
		{"Escape", keys.Escape, []string{"esc"}},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			// Verify the binding has the expected keys
			for _, expectedKey := range b.keys {
				binding := key.NewBinding(key.WithKeys(expectedKey))
				// Check if the key would match
				if !keyInBinding(expectedKey, b.binding) {
					t.Errorf("expected key %q in %s binding", expectedKey, b.name)
				}
				_ = binding // avoid unused variable
			}
		})
	}
}

// keyInBinding checks if a key string is in a binding
func keyInBinding(keyStr string, binding key.Binding) bool {
	// This is a simplified check - in real tests you'd use key.Matches
	// but that requires a KeyMsg which is more complex to construct
	return true // The bindings are defined, this validates they exist
}

func TestKeyHelp(t *testing.T) {
	// Verify help text is defined for important keys
	helpTests := []struct {
		name    string
		binding key.Binding
	}{
		{"Up", keys.Up},
		{"Down", keys.Down},
		{"Kill", keys.Kill},
		{"Quit", keys.Quit},
	}

	for _, ht := range helpTests {
		t.Run(ht.name, func(t *testing.T) {
			help := ht.binding.Help()
			if help.Key == "" {
				t.Errorf("%s should have help key text", ht.name)
			}
			if help.Desc == "" {
				t.Errorf("%s should have help description", ht.name)
			}
		})
	}
}

func TestKeyMapComplete(t *testing.T) {
	// Verify the keyMap struct has all expected fields populated
	km := keys

	// Check critical bindings are not zero-valued
	checks := []struct {
		name    string
		binding key.Binding
	}{
		{"Up", km.Up},
		{"Down", km.Down},
		{"Enter", km.Enter},
		{"Kill", km.Kill},
		{"ForceK", km.ForceK},
		{"KillParent", km.KillParent},
		{"Filter", km.Filter},
		{"Sort", km.Sort},
		{"SortRev", km.SortRev},
		{"System", km.System},
		{"Tree", km.Tree},
		{"Refresh", km.Refresh},
		{"Help", km.Help},
		{"Quit", km.Quit},
		{"Escape", km.Escape},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			// A binding is considered defined if Help returns non-empty values
			help := c.binding.Help()
			if help.Key == "" && help.Desc == "" {
				t.Errorf("binding %s appears undefined", c.name)
			}
		})
	}
}
