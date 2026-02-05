package tui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
)

type sortColumn int

const (
	sortByPort sortColumn = iota
	sortByPID
	sortByProcess
	sortByUser
	sortByMemory
	sortByUptime
	sortColumnCount
)

type sortState struct {
	column sortColumn
	asc    bool
}

type column struct {
	title string
	width int
}

// rowMeta holds per-row display metadata for tree rendering.
type rowMeta struct {
	treePrefix string // "", "├─ ", "└─ "
	isChild    bool
}

type portTable struct {
	columns   []column
	sort      sortState
	cfg       config.Config
	displayed []ports.PortInfo
	meta      []rowMeta // parallel to displayed
	treeMode  bool
	expanded  int // index of expanded row, -1 = none
	cursor    int
	offset    int
	height    int
	width     int
}

const prefixWidth = 2

func newPortTable(cfg config.Config) portTable {
	return portTable{
		columns: []column{
			{"PORT", 8},
			{"PID", 8},
			{"PROCESS", 20},
			{"USER", 12},
			{"MEMORY", 10},
			{"UPTIME", 10},
		},
		sort:     sortState{column: sortByPort, asc: true},
		cfg:      cfg,
		expanded: -1,
		treeMode: true,
	}
}

func (pt *portTable) setRows(items []ports.PortInfo) {
	sorted := make([]ports.PortInfo, len(items))
	copy(sorted, items)
	pt.sortItems(sorted)

	if pt.treeMode {
		sorted, pt.meta = buildTree(sorted)
	} else {
		pt.meta = make([]rowMeta, len(sorted))
	}
	pt.displayed = sorted

	if pt.cursor >= len(sorted) {
		pt.cursor = max(0, len(sorted)-1)
	}
	if pt.expanded >= len(sorted) {
		pt.expanded = -1
	}
	pt.clampScroll()
}

// buildTree reorders a sorted list so related processes are grouped together.
// Grouping rules (in priority order):
//  1. Same PID, different ports → group under first entry
//  2. PPID matches a PID in the list → child of that parent
//  3. Multiple items share a PPID not in the list → group as siblings
//  4. Everything else → standalone root
func buildTree(items []ports.PortInfo) ([]ports.PortInfo, []rowMeta) {
	// Step 1: group by PID (same process, multiple ports)
	type group struct {
		head     int   // index of first entry
		extraIdx []int // indices of additional port entries
	}
	pidFirst := make(map[int]int)  // PID -> index of first occurrence
	pidGroups := make(map[int]*group)
	order := make([]int, 0) // PIDs in order of first appearance
	for i, p := range items {
		if _, seen := pidFirst[p.PID]; !seen {
			pidFirst[p.PID] = i
			pidGroups[p.PID] = &group{head: i}
			order = append(order, p.PID)
		} else {
			pidGroups[p.PID].extraIdx = append(pidGroups[p.PID].extraIdx, i)
		}
	}

	// Step 2: determine parent-child among unique PIDs
	// A PID is a child if its PPID is another PID in the list
	childOf := make(map[int]int) // childPID -> parentPID
	for _, pid := range order {
		p := items[pidGroups[pid].head]
		if p.PPID > 1 && pidGroups[p.PPID] != nil {
			childOf[pid] = p.PPID
		}
	}

	// Step 3: group items with shared PPID not in the list
	// These become siblings grouped under the first one
	ppidPeers := make(map[int][]int) // PPID -> list of PIDs sharing it
	for _, pid := range order {
		if _, isChild := childOf[pid]; isChild {
			continue
		}
		p := items[pidGroups[pid].head]
		if p.PPID > 1 {
			ppidPeers[p.PPID] = append(ppidPeers[p.PPID], pid)
		}
	}
	// For shared PPIDs, first PID is root, rest are siblings
	siblingOf := make(map[int]int) // siblingPID -> rootPID
	for _, peers := range ppidPeers {
		if len(peers) < 2 {
			continue
		}
		root := peers[0]
		for _, pid := range peers[1:] {
			siblingOf[pid] = root
		}
	}

	// Step 4: collect children and siblings per root PID
	kids := make(map[int][]int) // rootPID -> child/sibling PIDs
	for pid, parent := range childOf {
		kids[parent] = append(kids[parent], pid)
	}
	for pid, root := range siblingOf {
		kids[root] = append(kids[root], pid)
	}

	// Step 5: build result — roots first, then their children
	result := make([]ports.PortInfo, 0, len(items))
	meta := make([]rowMeta, 0, len(items))

	appendItem := func(idx int, prefix string, isChild bool) {
		result = append(result, items[idx])
		meta = append(meta, rowMeta{treePrefix: prefix, isChild: isChild})
	}

	appendGroup := func(pid int, prefix string, isChild bool) {
		g := pidGroups[pid]
		appendItem(g.head, prefix, isChild)
		// Extra ports for same PID
		for ei, idx := range g.extraIdx {
			p := "│  ├─ "
			if ei == len(g.extraIdx)-1 && len(kids[pid]) == 0 {
				p = "│  └─ "
			}
			if !isChild {
				p = "├─ "
				if ei == len(g.extraIdx)-1 && len(kids[pid]) == 0 {
					p = "└─ "
				}
			}
			appendItem(idx, p, true)
		}
	}

	for _, pid := range order {
		// Skip if this PID is a child or sibling of another
		if _, isChild := childOf[pid]; isChild {
			continue
		}
		if _, isSibling := siblingOf[pid]; isSibling {
			continue
		}

		// Root entry
		appendGroup(pid, "", false)

		// Children and siblings
		allKids := kids[pid]
		for ci, kidPID := range allKids {
			prefix := "├─ "
			if ci == len(allKids)-1 {
				prefix = "└─ "
			}
			appendGroup(kidPID, prefix, true)
		}
	}

	return result, meta
}

func (pt *portTable) toggleExpand() {
	if pt.expanded == pt.cursor {
		pt.expanded = -1
	} else {
		pt.expanded = pt.cursor
	}
	pt.clampScroll()
}

func (pt *portTable) toggleTree() {
	pt.treeMode = !pt.treeMode
}

func (pt *portTable) rowLines(r int) int {
	if r == pt.expanded {
		return 1 + expandedLineCount(pt.displayed[r])
	}
	return 1
}

func (pt *portTable) view() string {
	var b strings.Builder

	// Header
	headerParts := []string{lipgloss.NewStyle().Width(prefixWidth).Render("")}
	for i, col := range pt.columns {
		title := col.title
		if sortColumn(i) == pt.sort.column {
			arrow := "▲"
			if !pt.sort.asc {
				arrow = "▼"
			}
			title += " " + arrow
		}
		headerParts = append(headerParts, tableHeaderStyle.Width(col.width).MaxWidth(col.width).Render(title))
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerParts...))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", pt.width))
	b.WriteString("\n")

	linesUsed := 0
	for r := pt.offset; r < len(pt.displayed) && linesUsed < pt.height; r++ {
		needed := pt.rowLines(r)
		if linesUsed+needed > pt.height {
			break
		}

		b.WriteString(pt.renderRow(r))
		linesUsed += needed

		if r == pt.expanded {
			b.WriteString(pt.renderExpanded(pt.displayed[r]))
		}

		if linesUsed < pt.height {
			b.WriteString("\n")
		}
	}

	for linesUsed < pt.height {
		b.WriteString("\n")
		linesUsed++
	}

	return b.String()
}

func (pt *portTable) renderRow(r int) string {
	p := pt.displayed[r]
	m := pt.meta[r]
	colorName := pt.cfg.PortColor(p.Port)
	pStyle := portStyle(colorName)

	prefix := "  "
	if r == pt.expanded {
		prefix = "▼ "
	} else if r == pt.cursor {
		prefix = "▸ "
	}

	// Tree-prefixed process name
	procName := p.Process
	if m.treePrefix != "" {
		procName = m.treePrefix + procName
	}

	// Dim style for child rows
	cStyle := cellStyle
	if m.isChild {
		cStyle = childCellStyle
	}

	cells := []string{
		lipgloss.NewStyle().Width(prefixWidth).Render(prefix),
		pStyle.Width(pt.columns[0].width).MaxWidth(pt.columns[0].width).Inline(true).Render(strconv.Itoa(p.Port)),
		cStyle.Width(pt.columns[1].width).MaxWidth(pt.columns[1].width).Inline(true).Render(strconv.Itoa(p.PID)),
		cStyle.Width(pt.columns[2].width).MaxWidth(pt.columns[2].width).Inline(true).Render(truncate(procName, pt.columns[2].width)),
		cStyle.Width(pt.columns[3].width).MaxWidth(pt.columns[3].width).Inline(true).Render(truncate(p.User, pt.columns[3].width)),
		cStyle.Width(pt.columns[4].width).MaxWidth(pt.columns[4].width).Inline(true).Render(p.Memory),
		cStyle.Width(pt.columns[5].width).MaxWidth(pt.columns[5].width).Inline(true).Render(p.Uptime),
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	if r == pt.cursor || r == pt.expanded {
		row = selectedRowStyle.Render(row)
	}
	return row
}

func (pt *portTable) renderExpanded(p ports.PortInfo) string {
	var lines []string

	pad := "      "
	labelW := 12

	add := func(label, value string) {
		if value == "" || value == "-" {
			return
		}
		l := expandLabelStyle.Width(labelW).Render(label)
		v := expandValueStyle.Render(value)
		lines = append(lines, pad+l+v)
	}

	add("Address", fmt.Sprintf("%s:%d", p.Address, p.Port))
	add("Command", p.Command)
	add("Directory", p.CWD)
	if p.Container != "" {
		add("Container", p.Container)
	}
	if p.PPID > 1 {
		l := expandLabelStyle.Width(labelW).Render("Parent PID")
		v := parentStyle.Render(strconv.Itoa(p.PPID))
		hint := expandLabelStyle.Render("  [p kill parent]")
		lines = append(lines, pad+l+v+hint)
	}

	return "\n" + strings.Join(lines, "\n")
}

func expandedLineCount(p ports.PortInfo) int {
	n := 2 // Address + Command
	if p.CWD != "" {
		n++
	}
	if p.Container != "" {
		n++
	}
	if p.PPID > 1 {
		n++
	}
	return n
}

func (pt *portTable) moveUp() {
	if pt.cursor > 0 {
		pt.cursor--
		pt.clampScroll()
	}
}

func (pt *portTable) moveDown() {
	if pt.cursor < len(pt.displayed)-1 {
		pt.cursor++
		pt.clampScroll()
	}
}

func (pt *portTable) clampScroll() {
	if pt.cursor < pt.offset {
		pt.offset = pt.cursor
	}
	lines := 0
	for r := pt.offset; r <= pt.cursor && r < len(pt.displayed); r++ {
		lines += pt.rowLines(r)
	}
	for lines > pt.height && pt.offset < pt.cursor {
		lines -= pt.rowLines(pt.offset)
		pt.offset++
	}
	if pt.offset < 0 {
		pt.offset = 0
	}
}

func (pt *portTable) sortItems(items []ports.PortInfo) {
	sort.Slice(items, func(i, j int) bool {
		var less bool
		switch pt.sort.column {
		case sortByPort:
			less = items[i].Port < items[j].Port
		case sortByPID:
			less = items[i].PID < items[j].PID
		case sortByProcess:
			less = items[i].Process < items[j].Process
		case sortByUser:
			less = items[i].User < items[j].User
		default:
			less = items[i].Port < items[j].Port
		}
		if !pt.sort.asc {
			return !less
		}
		return less
	})
}

func (pt *portTable) nextSort()    { pt.sort.column = (pt.sort.column + 1) % sortColumnCount; pt.sort.asc = true }
func (pt *portTable) reverseSort() { pt.sort.asc = !pt.sort.asc }

func (pt *portTable) setHeight(h int)  { pt.height = h }
func (pt *portTable) setWidth(w int)   { pt.width = w }
func (pt *portTable) selectedIndex() int { return pt.cursor }

func truncate(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		runes := []rune(s)
		if len(runes) > 0 {
			return string(runes[:1])
		}
		return ""
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > w-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func shortDir(cwd string) string {
	if cwd == "" {
		return "-"
	}
	base := filepath.Base(cwd)
	if base == "." || base == "/" {
		return cwd
	}
	return base
}
