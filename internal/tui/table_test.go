package tui

import (
	"strings"
	"testing"

	"github.com/legostin/reap/internal/config"
	"github.com/legostin/reap/internal/ports"
)

func TestNewPortTable(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	if len(pt.columns) != 6 {
		t.Errorf("expected 6 columns, got %d", len(pt.columns))
	}

	expectedColumns := []string{"PORT", "PID", "PROCESS", "USER", "MEMORY", "UPTIME"}
	for i, col := range pt.columns {
		if col.title != expectedColumns[i] {
			t.Errorf("column %d: expected %q, got %q", i, expectedColumns[i], col.title)
		}
	}

	if pt.sort.column != sortByPort {
		t.Errorf("expected default sort by port, got %d", pt.sort.column)
	}
	if !pt.sort.asc {
		t.Error("expected default ascending sort")
	}
	if pt.expanded != -1 {
		t.Errorf("expected expanded=-1, got %d", pt.expanded)
	}
	if !pt.treeMode {
		t.Error("expected treeMode=true by default")
	}
}

func TestPortTableSetRows(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.setHeight(20)
	pt.treeMode = false

	items := []ports.PortInfo{
		{Port: 8000, PID: 300, Process: "python"},
		{Port: 3000, PID: 100, Process: "node"},
		{Port: 5432, PID: 200, Process: "postgres"},
	}

	pt.setRows(items)

	if len(pt.displayed) != 3 {
		t.Fatalf("expected 3 displayed items, got %d", len(pt.displayed))
	}

	// Default sort is by port ascending
	if pt.displayed[0].Port != 3000 {
		t.Errorf("expected first item to be port 3000, got %d", pt.displayed[0].Port)
	}
	if pt.displayed[1].Port != 5432 {
		t.Errorf("expected second item to be port 5432, got %d", pt.displayed[1].Port)
	}
	if pt.displayed[2].Port != 8000 {
		t.Errorf("expected third item to be port 8000, got %d", pt.displayed[2].Port)
	}
}

func TestPortTableSortByPID(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.sort.column = sortByPID
	pt.sort.asc = true

	items := []ports.PortInfo{
		{Port: 3000, PID: 300},
		{Port: 5432, PID: 100},
		{Port: 8000, PID: 200},
	}

	pt.sortItems(items)

	if items[0].PID != 100 {
		t.Errorf("expected first PID=100, got %d", items[0].PID)
	}
	if items[1].PID != 200 {
		t.Errorf("expected second PID=200, got %d", items[1].PID)
	}
	if items[2].PID != 300 {
		t.Errorf("expected third PID=300, got %d", items[2].PID)
	}
}

func TestPortTableSortByProcess(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.sort.column = sortByProcess
	pt.sort.asc = true

	items := []ports.PortInfo{
		{Port: 3000, Process: "python"},
		{Port: 5432, Process: "node"},
		{Port: 8000, Process: "java"},
	}

	pt.sortItems(items)

	if items[0].Process != "java" {
		t.Errorf("expected first process=java, got %q", items[0].Process)
	}
	if items[1].Process != "node" {
		t.Errorf("expected second process=node, got %q", items[1].Process)
	}
	if items[2].Process != "python" {
		t.Errorf("expected third process=python, got %q", items[2].Process)
	}
}

func TestPortTableSortByUser(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.sort.column = sortByUser
	pt.sort.asc = true

	items := []ports.PortInfo{
		{Port: 3000, User: "user"},
		{Port: 5432, User: "admin"},
		{Port: 8000, User: "root"},
	}

	pt.sortItems(items)

	if items[0].User != "admin" {
		t.Errorf("expected first user=admin, got %q", items[0].User)
	}
	if items[1].User != "root" {
		t.Errorf("expected second user=root, got %q", items[1].User)
	}
	if items[2].User != "user" {
		t.Errorf("expected third user=user, got %q", items[2].User)
	}
}

func TestPortTableSortDescending(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.sort.column = sortByPort
	pt.sort.asc = false

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 5432},
		{Port: 8000},
	}

	pt.sortItems(items)

	if items[0].Port != 8000 {
		t.Errorf("expected first port=8000 (desc), got %d", items[0].Port)
	}
	if items[1].Port != 5432 {
		t.Errorf("expected second port=5432, got %d", items[1].Port)
	}
	if items[2].Port != 3000 {
		t.Errorf("expected third port=3000, got %d", items[2].Port)
	}
}

func TestPortTableNextSort(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	initial := pt.sort.column
	pt.nextSort()

	if pt.sort.column != (initial+1)%sortColumnCount {
		t.Errorf("expected sort column to cycle, got %d", pt.sort.column)
	}
	if !pt.sort.asc {
		t.Error("nextSort should reset to ascending")
	}
}

func TestPortTableReverseSort(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	if !pt.sort.asc {
		t.Error("initial sort should be ascending")
	}

	pt.reverseSort()
	if pt.sort.asc {
		t.Error("should be descending after reverseSort")
	}

	pt.reverseSort()
	if !pt.sort.asc {
		t.Error("should be ascending after second reverseSort")
	}
}

func TestPortTableToggleExpand(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node"},
		{Port: 5432, PID: 200, Process: "postgres"},
	}
	pt.setRows(items)

	if pt.expanded != -1 {
		t.Errorf("expected expanded=-1 initially, got %d", pt.expanded)
	}

	pt.cursor = 0
	pt.toggleExpand()
	if pt.expanded != 0 {
		t.Errorf("expected expanded=0 after toggle, got %d", pt.expanded)
	}

	pt.toggleExpand()
	if pt.expanded != -1 {
		t.Errorf("expected expanded=-1 after second toggle, got %d", pt.expanded)
	}
}

func TestPortTableToggleTree(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	if !pt.treeMode {
		t.Error("expected treeMode=true initially")
	}

	pt.toggleTree()
	if pt.treeMode {
		t.Error("expected treeMode=false after toggle")
	}

	pt.toggleTree()
	if !pt.treeMode {
		t.Error("expected treeMode=true after second toggle")
	}
}

func TestPortTableMoveUp(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 5432},
		{Port: 8000},
	}
	pt.setRows(items)

	pt.cursor = 2
	pt.moveUp()
	if pt.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", pt.cursor)
	}

	pt.moveUp()
	if pt.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", pt.cursor)
	}

	// Should not go below 0
	pt.moveUp()
	if pt.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", pt.cursor)
	}
}

func TestPortTableMoveDown(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 5432},
		{Port: 8000},
	}
	pt.setRows(items)

	pt.cursor = 0
	pt.moveDown()
	if pt.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", pt.cursor)
	}

	pt.moveDown()
	if pt.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", pt.cursor)
	}

	// Should not exceed list length
	pt.moveDown()
	if pt.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", pt.cursor)
	}
}

func TestPortTableSelectedIndex(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	pt.cursor = 5
	if pt.selectedIndex() != 5 {
		t.Errorf("expected selectedIndex=5, got %d", pt.selectedIndex())
	}
}

func TestPortTableSetHeight(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	pt.setHeight(50)
	if pt.height != 50 {
		t.Errorf("expected height=50, got %d", pt.height)
	}
}

func TestPortTableSetWidth(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	pt.setWidth(120)
	if pt.width != 120 {
		t.Errorf("expected width=120, got %d", pt.width)
	}
}

func TestPortTableRowLines(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000, PID: 100, Process: "node", Command: "node server.js", CWD: "/app"},
	}
	pt.setRows(items)

	// Non-expanded row is 1 line
	if pt.rowLines(0) != 1 {
		t.Errorf("expected 1 line for non-expanded row, got %d", pt.rowLines(0))
	}

	// Expanded row has additional lines
	pt.expanded = 0
	lines := pt.rowLines(0)
	if lines <= 1 {
		t.Errorf("expected >1 lines for expanded row, got %d", lines)
	}
}

func TestPortTableCursorClampOnSetRows(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 5432},
		{Port: 8000},
	}
	pt.setRows(items)
	pt.cursor = 2

	// Reduce items
	pt.setRows([]ports.PortInfo{{Port: 3000}})

	if pt.cursor != 0 {
		t.Errorf("cursor should be clamped to 0, got %d", pt.cursor)
	}
}

func TestPortTableExpandedClampOnSetRows(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 5432},
		{Port: 8000},
	}
	pt.setRows(items)
	pt.expanded = 2

	// Reduce items
	pt.setRows([]ports.PortInfo{{Port: 3000}})

	if pt.expanded != -1 {
		t.Errorf("expanded should be reset to -1, got %d", pt.expanded)
	}
}

func TestPortTableViewHeader(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setWidth(120)
	pt.setHeight(20)

	items := []ports.PortInfo{{Port: 3000, PID: 100, Process: "node", User: "user"}}
	pt.setRows(items)

	view := pt.view()

	// Check that header columns are present
	if !strings.Contains(view, "PORT") {
		t.Error("view should contain PORT header")
	}
	if !strings.Contains(view, "PID") {
		t.Error("view should contain PID header")
	}
	if !strings.Contains(view, "PROCESS") {
		t.Error("view should contain PROCESS header")
	}

	// Check sort indicator
	if !strings.Contains(view, "\u25b2") { // ▲
		t.Error("view should contain ascending sort indicator")
	}
}

func TestPortTableViewDescendingIndicator(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.sort.asc = false
	pt.setWidth(120)
	pt.setHeight(20)

	items := []ports.PortInfo{{Port: 3000}}
	pt.setRows(items)

	view := pt.view()

	if !strings.Contains(view, "\u25bc") { // ▼
		t.Error("view should contain descending sort indicator")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		width int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a long string", 10, "this is a\u2026"},
		{"abc", 1, "a"},
		{"", 10, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

func TestTruncateZeroWidth(t *testing.T) {
	// Based on implementation, width 0 returns first char if available
	got := truncate("abc", 0)
	// The implementation returns first rune for w <= 1
	if len(got) > 1 {
		t.Errorf("truncate with width 0 should return at most 1 char, got %q", got)
	}
}

func TestShortDir(t *testing.T) {
	tests := []struct {
		cwd  string
		want string
	}{
		{"", "-"},
		{"/", "/"},
		{"/home/user/projects/myapp", "myapp"},
		{"/usr/local/bin", "bin"},
		{".", "."},
	}

	for _, tt := range tests {
		got := shortDir(tt.cwd)
		if got != tt.want {
			t.Errorf("shortDir(%q) = %q, want %q", tt.cwd, got, tt.want)
		}
	}
}

func TestExpandedLineCount(t *testing.T) {
	tests := []struct {
		port ports.PortInfo
		min  int
		desc string
	}{
		{
			port: ports.PortInfo{Command: "node server.js"},
			min:  2,
			desc: "basic (address + command)",
		},
		{
			port: ports.PortInfo{Command: "node", CWD: "/app"},
			min:  3,
			desc: "with CWD",
		},
		{
			port: ports.PortInfo{Command: "node", Container: "my-container"},
			min:  3,
			desc: "with container",
		},
		{
			port: ports.PortInfo{Command: "node", PPID: 50},
			min:  3,
			desc: "with PPID",
		},
		{
			port: ports.PortInfo{Command: "node", CWD: "/app", Container: "my-container", PPID: 50},
			min:  5,
			desc: "all fields",
		},
	}

	for _, tt := range tests {
		got := expandedLineCount(tt.port)
		if got < tt.min {
			t.Errorf("%s: expandedLineCount got %d, want >= %d", tt.desc, got, tt.min)
		}
	}
}

func TestBuildTreeSamePID(t *testing.T) {
	// Same PID with multiple ports (e.g., Docker)
	items := []ports.PortInfo{
		{Port: 3000, PID: 100, PPID: 1, Process: "docker"},
		{Port: 5432, PID: 100, PPID: 1, Process: "docker"},
		{Port: 6379, PID: 100, PPID: 1, Process: "docker"},
	}

	result, meta := buildTree(items)

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}

	// First entry is root
	if meta[0].isChild {
		t.Error("first entry should be root")
	}

	// Extra entries are children
	if !meta[1].isChild || !meta[2].isChild {
		t.Error("extra port entries should be children")
	}
}

func TestBuildTreeParentChildPID(t *testing.T) {
	// PPID matches a PID in the list
	items := []ports.PortInfo{
		{Port: 9000, PID: 50, PPID: 1, Process: "php-fpm-master"},
		{Port: 9001, PID: 101, PPID: 50, Process: "php-fpm-worker"},
	}

	result, meta := buildTree(items)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	// Master is root
	if meta[0].isChild {
		t.Error("master should be root")
	}
	if result[0].PID != 50 {
		t.Error("master should be first")
	}

	// Worker is child
	if !meta[1].isChild {
		t.Error("worker should be child")
	}
}

func TestBuildTreeSharedPPID(t *testing.T) {
	// Processes share PPID not in list (siblings)
	items := []ports.PortInfo{
		{Port: 3000, PID: 101, PPID: 50, Process: "proxy1"},
		{Port: 3001, PID: 102, PPID: 50, Process: "proxy2"},
		{Port: 8080, PID: 200, PPID: 1, Process: "node"},
	}

	result, meta := buildTree(items)

	if len(result) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result))
	}

	// First proxy is root
	if meta[0].isChild {
		t.Error("first proxy should be root")
	}

	// Second proxy is sibling (grouped as child)
	if !meta[1].isChild {
		t.Error("second proxy should be grouped as child")
	}

	// Node is standalone
	if meta[2].isChild {
		t.Error("node should be standalone root")
	}
}

func TestBuildTreeEmpty(t *testing.T) {
	result, meta := buildTree([]ports.PortInfo{})

	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
	if len(meta) != 0 {
		t.Errorf("expected empty meta, got %d items", len(meta))
	}
}

func TestBuildTreeSingleItem(t *testing.T) {
	items := []ports.PortInfo{
		{Port: 3000, PID: 100, PPID: 1, Process: "node"},
	}

	result, meta := buildTree(items)

	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if meta[0].isChild {
		t.Error("single item should not be a child")
	}
}

func TestBuildTreePreservesOrder(t *testing.T) {
	items := []ports.PortInfo{
		{Port: 3000, PID: 100, PPID: 1, Process: "a"},
		{Port: 4000, PID: 200, PPID: 1, Process: "b"},
		{Port: 5000, PID: 300, PPID: 1, Process: "c"},
	}

	result, _ := buildTree(items)

	// All are roots with no relationship, order should be preserved
	if result[0].Port != 3000 || result[1].Port != 4000 || result[2].Port != 5000 {
		t.Errorf("order not preserved: %d, %d, %d", result[0].Port, result[1].Port, result[2].Port)
	}
}

func TestPortTableEmptyList(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.setWidth(120)
	pt.setHeight(20)

	pt.setRows([]ports.PortInfo{})

	if len(pt.displayed) != 0 {
		t.Errorf("expected empty displayed, got %d", len(pt.displayed))
	}

	// View should not panic
	view := pt.view()
	if view == "" {
		t.Error("view should not be empty even with no data")
	}
}

func TestSortColumnCycle(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)

	// Cycle through all sort columns
	seen := make(map[sortColumn]bool)
	for i := 0; i < int(sortColumnCount)+1; i++ {
		seen[pt.sort.column] = true
		pt.nextSort()
	}

	// Should have seen all columns
	for col := sortColumn(0); col < sortColumnCount; col++ {
		if !seen[col] {
			t.Errorf("sort column %d not seen in cycle", col)
		}
	}
}

func TestPortTableClampScrollBasic(t *testing.T) {
	cfg := config.Default()
	pt := newPortTable(cfg)
	pt.treeMode = false
	pt.setHeight(20)

	items := []ports.PortInfo{
		{Port: 3000},
		{Port: 4000},
		{Port: 5000},
	}
	pt.setRows(items)

	// Cursor at end, offset at 0
	pt.cursor = 2
	pt.offset = 0
	pt.clampScroll()

	// With height 20, all items should be visible, offset should stay 0
	if pt.offset != 0 {
		t.Errorf("offset should be 0, got %d", pt.offset)
	}
}
