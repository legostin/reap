package ports

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// enrichProcessInfo enriches PortInfo entries with uptime, memory, and full
// command by making a single ps call for all PIDs.
func enrichProcessInfo(ports []PortInfo) {
	if len(ports) == 0 {
		return
	}

	pidSet := make(map[int]bool)
	var pidArgs []string
	for _, p := range ports {
		if !pidSet[p.PID] {
			pidSet[p.PID] = true
			pidArgs = append(pidArgs, strconv.Itoa(p.PID))
		}
	}

	out, err := exec.Command("ps", "-o", "pid=,ppid=,etime=,rss=,command=", "-p", strings.Join(pidArgs, ",")).Output()
	if err != nil {
		return
	}

	type psInfo struct {
		ppid    int
		etime   string
		rss     int64 // KB
		command string
	}
	info := make(map[int]psInfo)

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		ppid, _ := strconv.Atoi(fields[1])
		rss, _ := strconv.ParseInt(fields[3], 10, 64)
		cmd := strings.Join(fields[4:], " ")
		info[pid] = psInfo{
			ppid:    ppid,
			etime:   fields[2],
			rss:     rss,
			command: cmd,
		}
	}

	for i := range ports {
		if ps, ok := info[ports[i].PID]; ok {
			ports[i].PPID = ps.ppid
			ports[i].Uptime = formatElapsed(ps.etime)
			ports[i].Memory = formatMemory(ps.rss)
			ports[i].Command = ps.command
		}
	}

	enrichCWD(ports)
}

// enrichCWD populates the CWD field using lsof -d cwd.
// Single call for all unique PIDs. Output format:
//
//	p<PID>
//	fcwd
//	n<path>
func enrichCWD(ports []PortInfo) {
	if len(ports) == 0 {
		return
	}

	pidSet := make(map[int]bool)
	var pidArgs []string
	for _, p := range ports {
		if !pidSet[p.PID] {
			pidSet[p.PID] = true
			pidArgs = append(pidArgs, strconv.Itoa(p.PID))
		}
	}

	out, err := exec.Command("lsof", "-a", "-d", "cwd", "-Fn",
		"-p", strings.Join(pidArgs, ",")).Output()
	if err != nil {
		return
	}

	cwdMap := parseLsofCWD(string(out))

	for i := range ports {
		if cwd, ok := cwdMap[ports[i].PID]; ok {
			ports[i].CWD = cwd
		}
	}
}

// parseLsofCWD parses lsof -Fn output for cwd file descriptors.
func parseLsofCWD(output string) map[int]string {
	m := make(map[int]string)
	var currentPID int

	for _, line := range strings.Split(output, "\n") {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(line[1:])
			if err == nil {
				currentPID = pid
			}
		case 'n':
			if currentPID > 0 {
				m[currentPID] = line[1:]
			}
		}
	}
	return m
}

// LookupProcess returns basic info about a PID (process name, command).
func LookupProcess(pid int) (name string, command string) {
	out, err := exec.Command("ps", "-o", "comm=,command=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", ""
	}
	line := strings.TrimSpace(string(out))
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", ""
	}
	name = fields[0]
	command = line
	return name, command
}

// formatElapsed converts ps etime format to human-readable.
// Formats: DD-HH:MM:SS, HH:MM:SS, MM:SS, SS
func formatElapsed(etime string) string {
	etime = strings.TrimSpace(etime)
	var days, hours, minutes, seconds int

	// Check for days: DD-...
	if idx := strings.Index(etime, "-"); idx != -1 {
		days, _ = strconv.Atoi(etime[:idx])
		etime = etime[idx+1:]
	}

	parts := strings.Split(etime, ":")
	switch len(parts) {
	case 3:
		hours, _ = strconv.Atoi(parts[0])
		minutes, _ = strconv.Atoi(parts[1])
		seconds, _ = strconv.Atoi(parts[2])
	case 2:
		minutes, _ = strconv.Atoi(parts[0])
		seconds, _ = strconv.Atoi(parts[1])
	case 1:
		seconds, _ = strconv.Atoi(parts[0])
	}

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// formatMemory converts KB to human-readable format.
func formatMemory(kb int64) string {
	if kb == 0 {
		return "-"
	}
	mb := float64(kb) / 1024.0
	if mb >= 1024 {
		return fmt.Sprintf("%.1f GB", mb/1024.0)
	}
	if mb >= 1 {
		return fmt.Sprintf("%.1f MB", mb)
	}
	return fmt.Sprintf("%d KB", kb)
}
