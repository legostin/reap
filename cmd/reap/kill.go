package main

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/legostin/reap/internal/ports"
	"github.com/spf13/cobra"
)

var (
	killForce bool
	killYes   bool
)

var killCmd = &cobra.Command{
	Use:   "kill <port>...",
	Short: "Kill processes on specified ports",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := ports.NewScanner()
		results, err := scanner.Scan()
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		portMap := make(map[int][]ports.PortInfo)
		for _, p := range results {
			portMap[p.Port] = append(portMap[p.Port], p)
		}

		for _, arg := range args {
			port, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid port: %s\n", arg)
				continue
			}

			procs, ok := portMap[port]
			if !ok {
				fmt.Fprintf(os.Stderr, "no process found on port %d\n", port)
				continue
			}

			for _, p := range procs {
				if !killYes {
					fmt.Printf("kill %s (PID %d) on port %d? [y/N] ", p.Process, p.PID, p.Port)
					var answer string
					fmt.Scanln(&answer)
					if answer != "y" && answer != "Y" {
						fmt.Println("skipped")
						continue
					}
				}

				sig := syscall.SIGTERM
				if killForce {
					sig = syscall.SIGKILL
				}

				if err := syscall.Kill(p.PID, sig); err != nil {
					fmt.Fprintf(os.Stderr, "failed to kill PID %d: %s\n", p.PID, err)
				} else {
					sigName := "SIGTERM"
					if killForce {
						sigName = "SIGKILL"
					}
					fmt.Printf("sent %s to %s (PID %d)\n", sigName, p.Process, p.PID)
				}
			}
		}
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killForce, "force", "f", false, "send SIGKILL instead of SIGTERM")
	killCmd.Flags().BoolVarP(&killYes, "yes", "y", false, "skip confirmation")
}
