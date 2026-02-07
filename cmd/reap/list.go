package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/legostin/reap/internal/ports"
	"github.com/spf13/cobra"
)

var (
	listPort   int
	listName   string
	listJSON   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List listening ports (non-interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := ports.NewScanner()
		results, err := scanner.Scan()
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		filtered := filterResults(results)

		if listJSON {
			return printJSON(filtered)
		}
		printTable(filtered)
		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(&listPort, "port", "p", 0, "filter by port number")
	listCmd.Flags().StringVarP(&listName, "name", "n", "", "filter by process name")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
}

func filterResults(results []ports.PortInfo) []ports.PortInfo {
	if listPort == 0 && listName == "" {
		return results
	}
	var filtered []ports.PortInfo
	for _, p := range results {
		if listPort != 0 && p.Port != listPort {
			continue
		}
		if listName != "" && !strings.Contains(strings.ToLower(p.Process), strings.ToLower(listName)) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

func printTable(items []ports.PortInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tPID\tPROCESS\tUSER\tMEMORY\tUPTIME\tCONTAINER\tDIR\tADDRESS")
	for _, p := range items {
		container := p.Container
		if container == "" {
			container = "-"
		}
		cwd := p.CWD
		if cwd == "" {
			cwd = "-"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s:%d\n",
			strconv.Itoa(p.Port), p.PID, p.Process, p.User, p.Memory, p.Uptime, container, cwd, p.Address, p.Port)
	}
	w.Flush()
}

func printJSON(items []ports.PortInfo) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
