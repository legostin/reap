package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
)

type mockScanner struct {
	ports []ports.PortInfo
	err   error
}

func (m *mockScanner) Scan() ([]ports.PortInfo, error) {
	return m.ports, m.err
}

func testModel() Model {
	scanner := &mockScanner{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user", Memory: "50.0 MB", Uptime: "1h 5m"},
			{Port: 5432, PID: 200, Process: "postgres", User: "user", Memory: "20.0 MB", Uptime: "2d 3h"},
			{Port: 8080, PID: 300, Process: "java", User: "root", Memory: "100.0 MB", Uptime: "30m 5s"},
		},
	}
	cfg := config.Default()
	m := New(scanner, cfg)
	// Simulate window size
	m.width = 120
	m.height = 40
	m.table.setWidth(120)
	m.table.setHeight(30)
	return m
}

func TestModelInit(t *testing.T) {
	m := testModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestPortsUpdated(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user"},
			{Port: 5432, PID: 200, Process: "postgres", User: "user"},
		},
	})
	model := updated.(Model)
	if len(model.allPorts) != 2 {
		t.Errorf("expected 2 ports, got %d", len(model.allPorts))
	}
	if len(model.filtered) != 2 {
		t.Errorf("expected 2 filtered ports, got %d", len(model.filtered))
	}
}

func TestQuitKey(t *testing.T) {
	m := testModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestFilterActivation(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model := updated.(Model)
	if !model.filter.active {
		t.Error("expected filter to be active after pressing /")
	}
}

func TestConfirmDialog(t *testing.T) {
	m := testModel()
	// Load ports first
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user"},
		},
	})
	m = updated.(Model)

	// Press k to kill
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)

	if !m.confirm.visible {
		t.Error("expected confirm dialog to be visible")
	}
	if m.confirm.force {
		t.Error("expected SIGTERM, not SIGKILL")
	}

	// Press n to cancel
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = updated.(Model)
	if m.confirm.visible {
		t.Error("expected confirm dialog to be hidden after n")
	}
}

func TestForceKillDialog(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user"},
		},
	})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	m = updated.(Model)

	if !m.confirm.visible {
		t.Error("expected confirm dialog to be visible")
	}
	if !m.confirm.force {
		t.Error("expected SIGKILL for force kill")
	}
}

func TestKillParent(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 9000, PID: 100, PPID: 50, Process: "php-fpm", User: "user"},
		},
	})
	m = updated.(Model)

	// Press p to kill parent
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = updated.(Model)

	if !m.confirm.visible {
		t.Error("expected confirm dialog for kill parent")
	}
	if !m.confirm.killParent {
		t.Error("expected killParent=true")
	}
}

func TestKillParentIgnoredForInit(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 9000, PID: 100, PPID: 1, Process: "php-fpm", User: "user"},
		},
	})
	m = updated.(Model)

	// Press p — should NOT show dialog because PPID=1 (init)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = updated.(Model)

	if m.confirm.visible {
		t.Error("should not show kill parent for PPID=1")
	}
}

func TestSortCycle(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user"},
		},
	})
	m = updated.(Model)

	initial := m.table.sort.column
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = updated.(Model)

	if m.table.sort.column == initial {
		t.Error("expected sort column to change after pressing s")
	}
}

func TestExpandToggle(t *testing.T) {
	m := testModel()
	updated, _ := m.Update(portsUpdatedMsg{
		ports: []ports.PortInfo{
			{Port: 3000, PID: 100, Process: "node", User: "user", Command: "node server.js", CWD: "/home/user/app"},
		},
	})
	m = updated.(Model)

	// Expand row
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.table.expanded != 0 {
		t.Errorf("expected expanded=0, got %d", m.table.expanded)
	}

	// Collapse with Enter again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.table.expanded != -1 {
		t.Errorf("expected expanded=-1, got %d", m.table.expanded)
	}

	// Expand and collapse with Esc
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	if m.table.expanded != -1 {
		t.Errorf("expected expanded=-1 after esc, got %d", m.table.expanded)
	}
}

func TestTreeGroupsSamePID(t *testing.T) {
	// Docker scenario: same PID listening on multiple ports
	items := []ports.PortInfo{
		{Port: 3000, PID: 100, PPID: 1, Process: "com.docker.backend"},
		{Port: 5432, PID: 100, PPID: 1, Process: "com.docker.backend"},
		{Port: 6379, PID: 100, PPID: 1, Process: "com.docker.backend"},
		{Port: 8080, PID: 200, PPID: 1, Process: "node"},
	}
	result, meta := buildTree(items)
	if len(result) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(result))
	}
	// First docker entry is root
	if meta[0].isChild {
		t.Error("first docker entry should be root")
	}
	// Extra docker ports are children
	if !meta[1].isChild || !meta[2].isChild {
		t.Error("extra docker ports should be children")
	}
	// node is standalone root
	if meta[3].isChild {
		t.Error("node should be root")
	}
}

func TestTreeGroupsSharedPPID(t *testing.T) {
	// Multiple processes share a parent not in the list
	items := []ports.PortInfo{
		{Port: 3000, PID: 101, PPID: 50, Process: "docker-proxy"},
		{Port: 5432, PID: 102, PPID: 50, Process: "docker-proxy"},
		{Port: 8080, PID: 200, PPID: 1, Process: "node"},
	}
	result, meta := buildTree(items)
	if len(result) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(result))
	}
	// First docker-proxy is root, second is sibling
	if meta[0].isChild {
		t.Error("first docker-proxy should be root")
	}
	if !meta[1].isChild {
		t.Error("second docker-proxy should be grouped as child")
	}
	if meta[2].isChild {
		t.Error("node should be root")
	}
	// Verify order preserved
	if result[0].Port != 3000 || result[1].Port != 5432 || result[2].Port != 8080 {
		t.Errorf("unexpected order: %d, %d, %d", result[0].Port, result[1].Port, result[2].Port)
	}
}

func TestTreeParentChild(t *testing.T) {
	// Child's PPID matches a parent PID in the list
	items := []ports.PortInfo{
		{Port: 9000, PID: 50, PPID: 1, Process: "php-fpm-master"},
		{Port: 9001, PID: 101, PPID: 50, Process: "php-fpm-worker"},
		{Port: 9002, PID: 102, PPID: 50, Process: "php-fpm-worker"},
	}
	result, meta := buildTree(items)
	if len(result) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(result))
	}
	if meta[0].isChild {
		t.Error("master should be root")
	}
	if !meta[1].isChild || !meta[2].isChild {
		t.Error("workers should be children")
	}
	if result[0].PID != 50 {
		t.Error("master should be first")
	}
}

func TestSystemToggle(t *testing.T) {
	m := testModel()
	initial := m.cfg.ShowSystem

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = updated.(Model)

	if m.cfg.ShowSystem == initial {
		t.Error("expected ShowSystem to toggle after pressing a")
	}
}
