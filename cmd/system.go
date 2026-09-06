package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// systemCmd is the parent command for system observability subcommands.
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System observability and resource monitoring",
	Long: `system provides subcommands for quick, human-readable system resource
and process visibility — CPU, memory, disk, top processes, and service logs.

Note: v1 wraps standard Linux tools (free, df, ps, journalctl, nproc) rather
than parsing /proc directly. A future version may add raw /proc parsing for
greater portability.`,
}

// --- nova system resources ---------------------------------------------------

var systemResourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Print CPU, memory, and disk usage",
	Long: `resources displays CPU count, memory usage (total/used/free), and disk
usage for the root filesystem. Wraps nproc, free -h, and df -h /.`,
	Run: runSystemResources,
}

func runSystemResources(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	if runtime.GOOS != "linux" {
		red.Println("Error: 'nova system resources' is only supported on Linux.")
		fmt.Printf("Detected OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	cyan.Println("========== System Resources ==========")

	// CPU count
	cyan.Println("\n--- CPU ---")
	runWrappedCommand("nproc", nil, "CPU count")

	// Memory
	cyan.Println("\n--- Memory ---")
	runWrappedCommand("free", []string{"-h"}, "memory info")

	// Disk
	cyan.Println("\n--- Disk (/) ---")
	runWrappedCommand("df", []string{"-h", "/"}, "disk info")

	cyan.Println("\n======================================")
}

// --- nova system processes ---------------------------------------------------

var systemProcessesCmd = &cobra.Command{
	Use:   "processes",
	Short: "List top processes by memory usage",
	Long: `processes lists the top 20 processes sorted by memory consumption.
Wraps 'ps aux --sort=-%mem'.`,
	Run: runSystemProcesses,
}

func runSystemProcesses(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	if runtime.GOOS != "linux" {
		red.Println("Error: 'nova system processes' is only supported on Linux.")
		fmt.Printf("Detected OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	cyan.Println("========== Top Processes (by memory) ==========")

	// ps aux --sort=-%mem — pipe through head is tricky with exec, so
	// capture output and print first 21 lines (header + 20 rows)
	psCmd := exec.Command("ps", "aux", "--sort=-%mem")
	out, err := psCmd.CombinedOutput()
	if err != nil {
		red.Printf("Error running ps: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")
	limit := 21 // header + 20 processes
	if len(lines) < limit {
		limit = len(lines)
	}
	for _, line := range lines[:limit] {
		fmt.Println(line)
	}

	cyan.Println("\n================================================")
}

// --- nova system logs [service] ----------------------------------------------

var systemLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show recent system or service logs",
	Long: `logs displays the most recent 50 log entries from journalctl. If a
service name is provided, only logs for that unit are shown.

Examples:
  nova system logs            — recent system logs
  nova system logs docker     — recent logs for the docker service
  nova system logs ssh        — recent logs for the ssh service`,
	Run: runSystemLogs,
}

func runSystemLogs(cmd *cobra.Command, args []string) {
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	if runtime.GOOS != "linux" {
		red.Println("Error: 'nova system logs' is only supported on Linux.")
		fmt.Printf("Detected OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	if _, err := exec.LookPath("journalctl"); err != nil {
		red.Println("Error: 'journalctl' not found in PATH. Is systemd available?")
		os.Exit(1)
	}

	var jArgs []string
	if len(args) > 0 {
		service := args[0]
		cyan.Printf("========== Logs: %s ==========\n", service)
		jArgs = []string{"-u", service, "--no-pager", "-n", "50"}
	} else {
		cyan.Println("========== Recent System Logs ==========")
		jArgs = []string{"--no-pager", "-n", "50"}
	}

	logCmd := exec.Command("journalctl", jArgs...)
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr

	if err := logCmd.Run(); err != nil {
		red.Printf("Error reading logs: %v\n", err)
		os.Exit(1)
	}

	cyan.Println("\n========================================")
}

// --- helpers -----------------------------------------------------------------

// runWrappedCommand runs a command and prints its output with a fallback error.
func runWrappedCommand(name string, args []string, label string) {
	red := color.New(color.FgRed)

	if _, err := exec.LookPath(name); err != nil {
		red.Printf("Error: '%s' not found in PATH (needed for %s)\n", name, label)
		return
	}

	wrapped := exec.Command(name, args...)
	wrapped.Stdout = os.Stdout
	wrapped.Stderr = os.Stderr

	if err := wrapped.Run(); err != nil {
		red.Printf("Error running '%s': %v\n", name, err)
	}
}

func init() {
	rootCmd.AddCommand(systemCmd)
	systemCmd.AddCommand(systemResourcesCmd)
	systemCmd.AddCommand(systemProcessesCmd)
	systemCmd.AddCommand(systemLogsCmd)
}
