package tui

import (
	"fmt"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
)

// Messages
type portsUpdatedMsg struct{ ports []ports.PortInfo }
type scanErrorMsg struct{ err error }
type tickMsg time.Time
type killResultMsg struct {
	pid int
	err error
}

type Model struct {
	scanner  ports.Scanner
	cfg      config.Config
	allPorts []ports.PortInfo
	filtered []ports.PortInfo
	table    portTable
	filter   filterInput
	confirm  confirmDialog
	width    int
	height   int
	scanning bool
	err      error
	status   string
	showHelp bool
}

func New(scanner ports.Scanner, cfg config.Config) Model {
	return Model{
		scanner: scanner,
		cfg:     cfg,
		table:   newPortTable(cfg),
		filter:  newFilterInput(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.scanner), tickCmd(m.cfg.RefreshInterval))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.setWidth(msg.Width)
		m.table.setHeight(msg.Height - 5) // header + status + padding
		return m, nil

	case portsUpdatedMsg:
		m.scanning = false
		m.allPorts = msg.ports
		m.applyFilter()
		m.status = fmt.Sprintf("%d ports", len(m.filtered))
		return m, nil

	case scanErrorMsg:
		m.scanning = false
		m.err = msg.err
		return m, nil

	case tickMsg:
		var cmd tea.Cmd
		if !m.scanning {
			m.scanning = true
			cmd = scanCmd(m.scanner)
		}
		return m, tea.Batch(cmd, tickCmd(m.cfg.RefreshInterval))

	case killResultMsg:
		if msg.err != nil {
			m.status = errorStyle.Render(fmt.Sprintf("kill failed: %s", msg.err))
		} else {
			m.status = successStyle.Render(fmt.Sprintf("killed PID %d", msg.pid))
		}
		m.scanning = true
		return m, scanCmd(m.scanner)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass to filter sub-component
	if m.filter.active {
		var cmd tea.Cmd
		m.filter.input, cmd = m.filter.input.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Confirm dialog has highest priority
	if m.confirm.visible {
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("y", "Y"))):
			target := m.confirm.target
			force := m.confirm.force
			pid := target.PID
			if m.confirm.killParent {
				pid = target.PPID
			}
			m.confirm.hide()
			return m, killCmd(pid, force)
		case key.Matches(msg, key.NewBinding(key.WithKeys("n", "N", "esc"))):
			m.confirm.hide()
			return m, nil
		}
		return m, nil
	}

	// Filter input
	if m.filter.active {
		switch {
		case key.Matches(msg, keys.Escape):
			m.filter.clear()
			m.applyFilter()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			m.filter.deactivate()
			return m, nil
		}
		var cmd tea.Cmd
		m.filter.input, cmd = m.filter.input.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	// Main keys
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Filter):
		m.filter.activate()
		return m, m.filter.input.Focus()
	case key.Matches(msg, keys.Kill):
		if target, ok := m.selectedPort(); ok {
			m.confirm.show(target, false, false)
		}
		return m, nil
	case key.Matches(msg, keys.ForceK):
		if target, ok := m.selectedPort(); ok {
			m.confirm.show(target, true, false)
		}
		return m, nil
	case key.Matches(msg, keys.KillParent):
		if target, ok := m.selectedPort(); ok {
			if target.PPID > 1 {
				m.confirm.show(target, false, true)
			}
		}
		return m, nil
	case key.Matches(msg, keys.Enter):
		m.table.toggleExpand()
		return m, nil
	case key.Matches(msg, keys.Escape):
		m.table.expanded = -1
		return m, nil
	case key.Matches(msg, keys.Sort):
		m.table.nextSort()
		m.table.setRows(m.filtered)
		return m, nil
	case key.Matches(msg, keys.SortRev):
		m.table.reverseSort()
		m.table.setRows(m.filtered)
		return m, nil
	case key.Matches(msg, keys.System):
		m.cfg.ShowSystem = !m.cfg.ShowSystem
		m.applyFilter()
		return m, nil
	case key.Matches(msg, keys.Tree):
		m.table.toggleTree()
		m.table.setRows(m.filtered)
		return m, nil
	case key.Matches(msg, keys.Refresh):
		m.scanning = true
		return m, scanCmd(m.scanner)
	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, keys.Up):
		m.table.moveUp()
		return m, nil
	case key.Matches(msg, keys.Down):
		m.table.moveDown()
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var sections []string

	// Header
	title := headerStyle.Width(m.width).Render(" reap — port monitor")
	sections = append(sections, title)

	// Filter bar
	if m.filter.active || m.filter.value() != "" {
		sections = append(sections, m.filter.input.View())
	}

	// Table
	sections = append(sections, m.table.view())

	// Status bar
	status := m.status
	if m.scanning {
		status = "scanning..."
	}
	if m.err != nil {
		status = errorStyle.Render(m.err.Error())
	}

	helpText := "↑↓ navigate  enter expand  k kill  K force  p kill parent  / filter  s/S sort  t tree  ? help  q quit"
	if m.showHelp {
		helpText = "↑/↓ navigate  j down  enter expand/collapse  k kill (SIGTERM)  K kill (SIGKILL)  " +
			"p kill parent  / filter  s sort column  S reverse  t toggle tree  a toggle system  r refresh  esc collapse  q quit\n" +
			"  " + colorLegend()
	}

	bar := statusBarStyle.Width(m.width).Render(
		fmt.Sprintf("%s  │  %s", status, helpText),
	)
	sections = append(sections, bar)

	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay confirm dialog
	if m.confirm.visible {
		dialog := m.confirm.view()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			dialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

	return view
}

func (m *Model) applyFilter() {
	m.filtered = nil
	for _, p := range m.allPorts {
		if !m.cfg.ShowSystem && isSystemProcess(p) {
			continue
		}
		if m.filter.matches(p) {
			m.filtered = append(m.filtered, p)
		}
	}
	m.table.setRows(m.filtered)
}

func (m Model) selectedPort() (ports.PortInfo, bool) {
	idx := m.table.selectedIndex()
	if idx < 0 || idx >= len(m.table.displayed) {
		return ports.PortInfo{}, false
	}
	return m.table.displayed[idx], true
}

func isSystemProcess(p ports.PortInfo) bool {
	system := []string{"launchd", "mDNSResponder", "bluetoothd", "rapportd",
		"sharingd", "ControlCenter", "SystemUIServer"}
	for _, s := range system {
		if p.Process == s {
			return true
		}
	}
	return p.PID <= 1
}

// Commands

func scanCmd(scanner ports.Scanner) tea.Cmd {
	return func() tea.Msg {
		results, err := scanner.Scan()
		if err != nil {
			return scanErrorMsg{err: err}
		}
		return portsUpdatedMsg{ports: results}
	}
}

func killCmd(pid int, force bool) tea.Cmd {
	return func() tea.Msg {
		sig := syscall.SIGTERM
		if force {
			sig = syscall.SIGKILL
		}
		err := syscall.Kill(pid, sig)
		return killResultMsg{pid: pid, err: err}
	}
}

func tickCmd(intervalSec int) tea.Cmd {
	return tea.Tick(time.Duration(intervalSec)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

