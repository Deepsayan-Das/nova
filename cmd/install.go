// TODO: revisit as part of Apollo (Phase 13 package manager)
// This is a thin wrapper around apt for Debian-based systems (GalactOS).
// It is deliberately NOT a package manager — no dependency resolution, no
// custom repositories, no package signing, no uninstall/rollback.

package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed install.yaml
var defaultInstallYAML []byte

// InstallTool represents a single entry in the install registry.
type InstallTool struct {
	Name        string   `json:"name" yaml:"name"`
	Key         string   `json:"key" yaml:"key"`
	Description string   `json:"description" yaml:"description"`
	Executables []string `json:"executables" yaml:"executables"`
	AptPackage  string   `json:"apt_package" yaml:"apt_package"`
}

// InstallConfig is the top-level structure of install.yaml.
type InstallConfig struct {
	Tools []InstallTool `json:"tools" yaml:"tools"`
}

// installCmd is the parent command for tool installation.
var installCmd = &cobra.Command{
	Use:   "install <tool>",
	Short: "Install a development tool via apt",
	Long: `install looks up the given tool name in Nova's embedded registry and
installs the corresponding apt package. This is a thin convenience wrapper for
Debian-based systems (GalactOS) — not a full package manager.

Use 'nova install list' to see all available tools and their install status.`,
	Args: cobra.MinimumNArgs(1),
	Run:  runInstall,
}

// installListCmd prints all registered tools and whether they are installed.
var installListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tools in the install registry and their install status",
	Run:   runInstallList,
}

func runInstall(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	if runtime.GOOS != "linux" {
		red.Println("Error: 'nova install' is only supported on Linux (GalactOS is Debian-based).")
		fmt.Printf("Detected OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	toolKey := strings.ToLower(args[0])

	tools := loadInstallRegistry()
	var target *InstallTool
	for i, t := range tools {
		if strings.EqualFold(t.Key, toolKey) || strings.EqualFold(t.Name, toolKey) {
			target = &tools[i]
			break
		}
	}

	if target == nil {
		red.Printf("Error: unknown tool '%s'\n", toolKey)
		fmt.Println("Run 'nova install list' to see available tools.")
		os.Exit(1)
	}

	// Check if already installed
	allInstalled := true
	for _, exe := range target.Executables {
		if _, err := exec.LookPath(exe); err != nil {
			allInstalled = false
			break
		}
	}
	if allInstalled {
		green.Printf("✓ %s is already installed\n", target.Name)
		return
	}

	cyan.Printf("Installing %s (apt package: %s)...\n", target.Name, target.AptPackage)

	// Check that apt exists
	if _, err := exec.LookPath("apt"); err != nil {
		red.Println("Error: 'apt' not found in PATH. Is this a Debian-based system?")
		os.Exit(1)
	}

	// Check that sudo exists
	if _, err := exec.LookPath("sudo"); err != nil {
		yellow.Println("Warning: 'sudo' not found — attempting install without sudo...")
		runAptInstall("apt", target.AptPackage)
		return
	}

	runAptInstall("sudo", target.AptPackage)

	// Verify
	allInstalled = true
	for _, exe := range target.Executables {
		if _, err := exec.LookPath(exe); err != nil {
			allInstalled = false
		}
	}
	if allInstalled {
		green.Printf("✓ %s installed successfully\n", target.Name)
	} else {
		yellow.Printf("⚠ apt completed but some executables for %s are still not in PATH\n", target.Name)
	}
}

func runAptInstall(sudoOrApt string, aptPackage string) {
	red := color.New(color.FgRed)

	var aptCmd *exec.Cmd
	if sudoOrApt == "sudo" {
		aptCmd = exec.Command("sudo", "apt", "install", "-y", aptPackage)
	} else {
		aptCmd = exec.Command("apt", "install", "-y", aptPackage)
	}
	aptCmd.Stdout = os.Stdout
	aptCmd.Stderr = os.Stderr
	aptCmd.Stdin = os.Stdin

	if err := aptCmd.Run(); err != nil {
		red.Printf("Error: install failed: %v\n", err)
		os.Exit(1)
	}
}

func runInstallList(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	tools := loadInstallRegistry()

	cyan.Println("========== Nova Install Registry ==========")
	for _, t := range tools {
		installed := true
		for _, exe := range t.Executables {
			if _, err := exec.LookPath(exe); err != nil {
				installed = false
				break
			}
		}
		status := green.Sprintf("✓ installed")
		if !installed {
			status = red.Sprintf("✗ not found")
		}
		fmt.Printf("  %-14s %s  —  %s\n", t.Key, status, t.Description)
	}
	cyan.Println("============================================")
}

func loadInstallRegistry() []InstallTool {
	red := color.New(color.FgRed)

	var config InstallConfig
	if err := yaml.Unmarshal(defaultInstallYAML, &config); err != nil {
		red.Printf("Error parsing install.yaml: %v\n", err)
		os.Exit(1)
	}
	return config.Tools
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(installListCmd)
}
