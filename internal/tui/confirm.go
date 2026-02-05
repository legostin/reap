package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/legostin/reap/internal/ports"
)

type confirmDialog struct {
	visible    bool
	target     ports.PortInfo
	force      bool
	killParent bool
}

func (d *confirmDialog) show(target ports.PortInfo, force bool, killParent bool) {
	d.visible = true
	d.target = target
	d.force = force
	d.killParent = killParent
}

func (d *confirmDialog) hide() {
	d.visible = false
}

func (d *confirmDialog) view() string {
	if !d.visible {
		return ""
	}

	signal := "SIGTERM"
	if d.force {
		signal = "SIGKILL"
	}

	var title, body string

	if d.killParent {
		title = dialogTitleStyle.Render(fmt.Sprintf("Kill PARENT process? (%s)", signal))
		body = fmt.Sprintf(
			"\n  Target:  %s (PID %d, port %d)"+
				"\n  Parent:  PID %d"+
				"\n\n  This will kill the parent and all its children.\n",
			d.target.Process, d.target.PID, d.target.Port, d.target.PPID,
		)
	} else {
		title = dialogTitleStyle.Render(fmt.Sprintf("Kill process? (%s)", signal))
		body = fmt.Sprintf(
			"\n  Process: %s\n  PID:     %d\n  Port:    %d\n  User:    %s\n",
			d.target.Process, d.target.PID, d.target.Port, d.target.User,
		)
	}

	prompt := "\n  " + lipgloss.NewStyle().Bold(true).Render("y") + " confirm  " +
		lipgloss.NewStyle().Bold(true).Render("n/esc") + " cancel"

	return dialogStyle.Render(title + body + prompt)
}
