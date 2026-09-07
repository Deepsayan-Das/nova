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

// installCmd is the parent command for tool and package installation.
var installCmd = &cobra.Command{
	Use:   "install <package...>",
	Short: "Install system packages via apt (rebranded package manager)",
	Long: `install is GalactOS's rebranded package installation command powered by apt under the hood.

If the package matches a shortcut in Nova's tool registry (e.g. 'go', 'java', 'docker'),
it automatically resolves to the corresponding apt package name (e.g. 'golang', 'default-jdk', 'docker.io').
Otherwise, the package name is passed directly to 'apt install'.

Examples:
  nova install go            — installs golang via apt
  nova install htop          — installs htop directly via apt
  nova install git curl vim  — installs multiple packages via apt
  nova install list          — list predefined tool shortcuts and install status`,
	Args: cobra.MinimumNArgs(1),
	Run:  runInstall,
}

// installListCmd prints all registered tools and whether they are installed.
var installListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tools in the install registry and their install status",
	Run:   runInstallList,
}

type installTarget struct {
	Name        string
	AptPackage  string
	Executables []string
}

// resolveAptPackage resolves a tool key or package name to its target apt package and metadata.
func resolveAptPackage(arg string, tools []InstallTool) (displayName string, aptPackage string, executables []string) {
	toolKey := strings.ToLower(arg)
	for _, t := range tools {
		if strings.EqualFold(t.Key, toolKey) || strings.EqualFold(t.Name, toolKey) {
			return t.Name, t.AptPackage, t.Executables
		}
	}
	return arg, arg, []string{arg}
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

	tools := loadInstallRegistry()

	var targets []installTarget
	var aptPackages []string

	for _, arg := range args {
		name, aptPkg, exes := resolveAptPackage(arg, tools)

		allInstalled := len(exes) > 0
		for _, exe := range exes {
			if _, err := exec.LookPath(exe); err != nil {
				allInstalled = false
				break
			}
		}

		if allInstalled {
			green.Printf("✓ %s is already installed\n", name)
			continue
		}

		targets = append(targets, installTarget{
			Name:        name,
			AptPackage:  aptPkg,
			Executables: exes,
		})
		aptPackages = append(aptPackages, aptPkg)
	}

	if len(aptPackages) == 0 {
		return
	}

	// Check that apt exists
	if _, err := exec.LookPath("apt"); err != nil {
		red.Println("Error: 'apt' not found in PATH. Is this a Debian-based system?")
		os.Exit(1)
	}

	cyan.Printf("Installing package(s) via apt: %s...\n", strings.Join(aptPackages, ", "))

	// Check that sudo exists
	if _, err := exec.LookPath("sudo"); err != nil {
		yellow.Println("Warning: 'sudo' not found — attempting install without sudo...")
		runAptInstall("apt", aptPackages)
	} else {
		runAptInstall("sudo", aptPackages)
	}

	// Verify
	for _, target := range targets {
		allInstalled := len(target.Executables) > 0
		for _, exe := range target.Executables {
			if _, err := exec.LookPath(exe); err != nil {
				allInstalled = false
				break
			}
		}
		if allInstalled {
			green.Printf("✓ %s installed successfully\n", target.Name)
		} else {
			yellow.Printf("⚠ apt completed, but executable(s) for %s are still not in PATH\n", target.Name)
		}
	}
}

func runAptInstall(sudoOrApt string, aptPackages []string) {
	red := color.New(color.FgRed)

	aptArgs := append([]string{"install", "-y"}, aptPackages...)

	var aptCmd *exec.Cmd
	if sudoOrApt == "sudo" {
		cmdArgs := append([]string{"apt"}, aptArgs...)
		aptCmd = exec.Command("sudo", cmdArgs...)
	} else {
		aptCmd = exec.Command("apt", aptArgs...)
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
