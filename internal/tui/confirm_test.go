package tui

import (
	"strings"
	"testing"

	"github.com/legostin/reap/internal/ports"
)

func TestConfirmDialogInitial(t *testing.T) {
	d := confirmDialog{}

	if d.visible {
		t.Error("dialog should not be visible initially")
	}
	if d.force {
		t.Error("force should be false initially")
	}
	if d.killParent {
		t.Error("killParent should be false initially")
	}
}

func TestConfirmDialogShow(t *testing.T) {
	d := confirmDialog{}
	target := ports.PortInfo{
		Port:    3000,
		PID:     1234,
		PPID:    100,
		Process: "node",
		User:    "testuser",
	}

	d.show(target, false, false)

	if !d.visible {
		t.Error("dialog should be visible after show()")
	}
	if d.force {
		t.Error("force should be false for SIGTERM")
	}
	if d.killParent {
		t.Error("killParent should be false")
	}
	if d.target.Port != 3000 {
		t.Errorf("expected target port=3000, got %d", d.target.Port)
	}
	if d.target.PID != 1234 {
		t.Errorf("expected target PID=1234, got %d", d.target.PID)
	}
	if d.target.Process != "node" {
		t.Errorf("expected target process=node, got %q", d.target.Process)
	}
}

func TestConfirmDialogShowForceKill(t *testing.T) {
	d := confirmDialog{}
	target := ports.PortInfo{
		Port:    3000,
		PID:     1234,
		Process: "node",
	}

	d.show(target, true, false)

	if !d.visible {
		t.Error("dialog should be visible")
	}
	if !d.force {
		t.Error("force should be true for SIGKILL")
	}
}

func TestConfirmDialogShowKillParent(t *testing.T) {
	d := confirmDialog{}
	target := ports.PortInfo{
		Port:    3000,
		PID:     1234,
		PPID:    100,
		Process: "node",
	}

	d.show(target, false, true)

	if !d.visible {
		t.Error("dialog should be visible")
	}
	if !d.killParent {
		t.Error("killParent should be true")
	}
	if d.target.PPID != 100 {
		t.Errorf("expected PPID=100, got %d", d.target.PPID)
	}
}

func TestConfirmDialogShowForceKillParent(t *testing.T) {
	d := confirmDialog{}
	target := ports.PortInfo{
		Port:    3000,
		PID:     1234,
		PPID:    100,
		Process: "node",
	}

	d.show(target, true, true)

	if !d.force {
		t.Error("force should be true")
	}
	if !d.killParent {
		t.Error("killParent should be true")
	}
}

func TestConfirmDialogHide(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{Port: 3000}, false, false)

	d.hide()

	if d.visible {
		t.Error("dialog should not be visible after hide()")
	}
}

func TestConfirmDialogViewHidden(t *testing.T) {
	d := confirmDialog{}
	d.visible = false

	view := d.view()

	if view != "" {
		t.Errorf("hidden dialog should return empty view, got %q", view)
	}
}

func TestConfirmDialogViewSIGTERM(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{
		Port:    3000,
		PID:     1234,
		Process: "node",
		User:    "testuser",
	}, false, false)

	view := d.view()

	if !strings.Contains(view, "SIGTERM") {
		t.Error("view should mention SIGTERM")
	}
	if strings.Contains(view, "SIGKILL") {
		t.Error("view should not mention SIGKILL for normal kill")
	}
	if !strings.Contains(view, "Kill process?") {
		t.Error("view should ask 'Kill process?'")
	}
	if !strings.Contains(view, "node") {
		t.Error("view should show process name")
	}
	if !strings.Contains(view, "1234") {
		t.Error("view should show PID")
	}
	if !strings.Contains(view, "3000") {
		t.Error("view should show port")
	}
	if !strings.Contains(view, "testuser") {
		t.Error("view should show user")
	}
}

func TestConfirmDialogViewSIGKILL(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{
		Port:    3000,
		PID:     1234,
		Process: "node",
	}, true, false)

	view := d.view()

	if !strings.Contains(view, "SIGKILL") {
		t.Error("view should mention SIGKILL for force kill")
	}
	if strings.Contains(view, "SIGTERM") {
		t.Error("view should not mention SIGTERM for force kill")
	}
}

func TestConfirmDialogViewKillParent(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{
		Port:    3000,
		PID:     1234,
		PPID:    100,
		Process: "php-fpm",
	}, false, true)

	view := d.view()

	if !strings.Contains(view, "Kill PARENT process?") {
		t.Error("view should ask about killing parent")
	}
	if !strings.Contains(view, "100") {
		t.Error("view should show parent PID")
	}
	if !strings.Contains(view, "children") {
		t.Error("view should mention killing children")
	}
}

func TestConfirmDialogViewPrompt(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{Port: 3000, PID: 1234, Process: "node"}, false, false)

	view := d.view()

	if !strings.Contains(view, "y") {
		t.Error("view should show 'y' for confirm")
	}
	if !strings.Contains(view, "confirm") {
		t.Error("view should show 'confirm'")
	}
	if !strings.Contains(view, "n") || !strings.Contains(view, "esc") {
		t.Error("view should show 'n/esc' for cancel")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("view should show 'cancel'")
	}
}

func TestConfirmDialogPreservesTarget(t *testing.T) {
	d := confirmDialog{}

	target1 := ports.PortInfo{Port: 3000, PID: 100, Process: "node"}
	d.show(target1, false, false)

	if d.target.Port != 3000 {
		t.Error("target should be preserved")
	}

	// Show different target
	target2 := ports.PortInfo{Port: 5432, PID: 200, Process: "postgres"}
	d.show(target2, true, false)

	if d.target.Port != 5432 {
		t.Errorf("target should be updated to 5432, got %d", d.target.Port)
	}
	if d.target.Process != "postgres" {
		t.Errorf("target process should be postgres, got %q", d.target.Process)
	}
}

func TestConfirmDialogHidePreservesTarget(t *testing.T) {
	d := confirmDialog{}
	target := ports.PortInfo{Port: 3000, PID: 1234, Process: "node"}
	d.show(target, false, false)
	d.hide()

	// Target should still be accessible after hide
	if d.target.Port != 3000 {
		t.Error("target should be preserved after hide")
	}
}

func TestConfirmDialogForceKillParent(t *testing.T) {
	d := confirmDialog{}
	d.show(ports.PortInfo{
		Port:    9000,
		PID:     1234,
		PPID:    100,
		Process: "php-fpm",
	}, true, true)

	view := d.view()

	if !strings.Contains(view, "SIGKILL") {
		t.Error("force kill parent should show SIGKILL")
	}
	if !strings.Contains(view, "Kill PARENT") {
		t.Error("should indicate killing parent")
	}
}
