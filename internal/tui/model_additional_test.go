package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
)

func TestIsSystemProcess(t *testing.T) {
	tests := []struct {
		port   ports.PortInfo
		system bool
		desc   string
	}{
		{ports.PortInfo{Process: "launchd", PID: 1}, true, "launchd"},
		{ports.PortInfo{Process: "mDNSResponder", PID: 100}, true, "mDNSResponder"},
		{ports.PortInfo{Process: "bluetoothd", PID: 100}, true, "bluetoothd"},
		{ports.PortInfo{Process: "rapportd", PID: 100}, true, "rapportd"},
		{ports.PortInfo{Process: "sharingd", PID: 100}, true, "sharingd"},
		{ports.PortInfo{Process: "ControlCenter", PID: 100}, true, "ControlCenter"},
		{ports.PortInfo{Process: "SystemUIServer", PID: 100}, true, "SystemUIServer"},
		{ports.PortInfo{Process: "node", PID: 100}, false, "node"},
		{ports.PortInfo{Process: "python", PID: 100}, false, "python"},
		{ports.PortInfo{Process: "postgres", PID: 100}, false, "postgres"},
		{ports.PortInfo{Process: "anything", PID: 1}, true, "PID 1"},
		{ports.PortInfo{Process: "anything", PID: 0}, true, "PID 0"},
	}

	for _, tt := range tests {
		got := isSystemProcess(tt.port)
		if got != tt.system {
			t.Errorf("%s: isSystemProcess() = %v, want %v", tt.desc, got, tt.system)
		}
	}
}

func TestModelNew(t *testing.T) {
	scanner := &mockScanner{ports: []ports.PortInfo{}}
	cfg := config.Default()
	m := New(scanner, cfg)

	if m.scanner == nil {
		t.Error("scanner should not be nil")
	}
	if m.filter.active {
		t.Error("filter should not be active initially")
	}
	if m.confirm.visible {
		t.Error("confirm dialog should not be visible initially")
	}
	if m.showHelp {
		t.Error("help should not be shown initially")
	}
}

func TestModelWindowSize(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	model := updated.(Model)

	if model.width != 200 {
		t.Errorf("expected width=200, got %d", model.width)
	}
	if model.height != 50 {
		t.Errorf("expected height=50, got %d", model.height)
	}
}

func TestModelScanError(t *testing.T) {
	m := testModel()
	testErr := &testError{msg: "scan failed"}
	updated, _ := m.Update(scanErrorMsg{err: testErr})
	model := updated.(Model)

	if model.err == nil {
		t.Error("expected error to be set")
	}
	if model.err.Error() != "scan failed" {
		t.Errorf("expected 'scan failed', got %q", model.err.Error())
	}
	if model.scanning {
		t.Error("scanning should be false after error")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestModelTickWhileScanning(t *testing.T) {
	m := testModel()
	m.scanning = true

	updated, cmd := m.Update(tickMsg{})
	model := updated.(Model)

	// Should still be scanning, shouldn't trigger new scan
	if !model.scanning {
		t.Error("should remain scanning")
	}
	// cmd should include tick but not new scan
	if cmd == nil {
		t.Error("expected tick command to return")
	}
}

func TestModelTickWhileNotScanning(t *testing.T) {
	m := testModel()
	m.scanning = false

	updated, cmd := m.Update(tickMsg{})
	model := updated.(Model)

	// Should trigger new scan
	if !model.scanning {
		t.Error("should start scanning")
	}
	if cmd == nil {
		t.Error("expected command to return")
	}
}

func TestModelRefresh(t *testing.T) {
	m := testModel()
	m.scanning = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model := updated.(Model)

	if !model.scanning {
		t.Error("r key should trigger scanning")
	}
	if cmd == nil {
		t.Error("expected scan command")
	}
}

func TestModelHelpToggle(t *testing.T) {
	m := testModel()

	if m.showHelp {
		t.Error("help should be hidden initially")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model := updated.(Model)

	if !model.showHelp {
		t.Error("help should be shown after ?")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model = updated.(Model)

	if model.showHelp {
		t.Error("help should be hidden after second ?")
	}
}

func TestModelNavigateUp(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "a"},
			{Port: 4000, PID: 200, Process: "b"},
			{Port: 5000, PID: 300, Process: "c"},
		},
	})
	m = updated.(Model)
	m.table.cursor = 2

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(Model)

	if model.table.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", model.table.cursor)
	}
}

func TestModelNavigateDown(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "a"},
			{Port: 4000, PID: 200, Process: "b"},
		},
	})
	m = updated.(Model)
	m.table.cursor = 0

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model := updated.(Model)

	if model.table.cursor != 1 {
		t.Errorf("expected cursor=1 after j, got %d", model.table.cursor)
	}
}

func TestModelTreeToggle(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000}},
	})
	m = updated.(Model)

	initial := m.table.treeMode
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	model := updated.(Model)

	if model.table.treeMode == initial {
		t.Error("t should toggle tree mode")
	}
}

func TestModelSortReverse(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000}},
	})
	m = updated.(Model)

	initial := m.table.sort.asc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")})
	model := updated.(Model)

	if model.table.sort.asc == initial {
		t.Error("S should reverse sort direction")
	}
}

func TestModelEscapeCollapsesExpanded(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000, Command: "node"}},
	})
	m = updated.(Model)

	// Expand
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.table.expanded != 0 {
		t.Error("row should be expanded")
	}

	// Escape collapses
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)
	if model.table.expanded != -1 {
		t.Error("esc should collapse expanded row")
	}
}

func TestModelFilterEscapeClearsFilter(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, Process: "node"},
			{Port: 5432, Process: "postgres"},
		},
	})
	m = updated.(Model)

	// Activate filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updated.(Model)
	m.filter.input.SetValue("node")
	m.applyFilter()

	// Escape clears filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	if model.filter.active {
		t.Error("filter should be inactive after esc")
	}
	if model.filter.value() != "" {
		t.Error("filter value should be cleared")
	}
}

func TestModelFilterEnterDeactivates(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000, Process: "node"}},
	})
	m = updated.(Model)

	// Activate filter
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = updated.(Model)
	m.filter.input.SetValue("node")

	// Enter deactivates but keeps value
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.filter.active {
		t.Error("filter should be inactive after enter")
	}
	if model.filter.value() != "node" {
		t.Error("filter value should be preserved")
	}
}

func TestModelConfirmDialogEscCancel(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000, PID: 100, Process: "node"}},
	})
	m = updated.(Model)

	// Show kill dialog
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)

	// Cancel with escape
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model := updated.(Model)

	if model.confirm.visible {
		t.Error("esc should hide confirm dialog")
	}
}

func TestModelViewLoading(t *testing.T) {
	m := testModel()
	m.width = 0 // No window size yet

	view := m.View()
	if view != "loading..." {
		t.Errorf("expected 'loading...', got %q", view)
	}
}

func TestModelViewWithData(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{{Port: 3000, PID: 100, Process: "node", User: "user"}},
	})
	model := updated.(Model)

	view := model.View()
	if view == "loading..." {
		t.Error("should not show loading when data is present")
	}
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}

func TestModelKillResult(t *testing.T) {
	m := testModel()
	updated, cmd := m.Update(killResultMsg{pid: 1234, err: nil})
	model := updated.(Model)

	if model.status == "" {
		t.Error("status should be updated after kill")
	}
	if !model.scanning {
		t.Error("should trigger rescan after kill")
	}
	if cmd == nil {
		t.Error("expected scan command after kill")
	}
}

func TestModelKillResultError(t *testing.T) {
	m := testModel()
	testErr := &testError{msg: "permission denied"}
	updated, _ := m.Update(killResultMsg{pid: 1234, err: testErr})
	model := updated.(Model)

	if model.status == "" {
		t.Error("status should be updated after kill error")
	}
}

func TestModelSelectedPortEmpty(t *testing.T) {
	m := testModel()
	m.table.displayed = nil

	port, ok := m.selectedPort()
	if ok {
		t.Error("should return false when no ports")
	}
	if port.Port != 0 {
		t.Error("should return zero port")
	}
}

func TestModelSelectedPortOutOfBounds(t *testing.T) {
	m := testModel()
	m.table.displayed = []ports.PortInfo{{Port: 3000}}
	m.table.cursor = 5 // Out of bounds

	_, ok := m.selectedPort()
	if ok {
		t.Error("should return false when cursor out of bounds")
	}
}

func TestModelApplyFilterWithSystem(t *testing.T) {
	m := testModel()
	m.allPorts = []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node"},
		{Port: 5000, PID: 1, Process: "launchd"},
	}

	// Without show system
	m.cfg.ShowSystem = false
	m.applyFilter()
	if len(m.filtered) != 1 {
		t.Errorf("expected 1 port without system, got %d", len(m.filtered))
	}

	// With show system
	m.cfg.ShowSystem = true
	m.applyFilter()
	if len(m.filtered) != 2 {
		t.Errorf("expected 2 ports with system, got %d", len(m.filtered))
	}
}

func TestModelCtrlCQuit(t *testing.T) {
	m := testModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
}
