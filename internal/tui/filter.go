package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/legostin/reap/internal/ports"
)

type filterInput struct {
	input  textinput.Model
	active bool
}

func newFilterInput() filterInput {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.PromptStyle = filterPromptStyle
	ti.CharLimit = 64
	return filterInput{input: ti}
}

func (f *filterInput) activate() {
	f.active = true
	f.input.Focus()
}

func (f *filterInput) deactivate() {
	f.active = false
	f.input.Blur()
}

func (f *filterInput) clear() {
	f.input.SetValue("")
	f.deactivate()
}

func (f *filterInput) value() string {
	return f.input.Value()
}

func (f *filterInput) matches(p ports.PortInfo) bool {
	q := strings.ToLower(f.input.Value())
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(p.Process), q) ||
		strings.Contains(strconv.Itoa(p.Port), q) ||
		strings.Contains(strconv.Itoa(p.PID), q) ||
		strings.Contains(strings.ToLower(p.User), q) ||
		strings.Contains(strings.ToLower(p.Container), q) ||
		strings.Contains(strings.ToLower(p.CWD), q)
}
