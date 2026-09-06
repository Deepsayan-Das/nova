// nova update — source-based self-update.
//
// Known limitation: this only works when running from the nova source
// directory (a git clone). A proper update mechanism — checking a version
// endpoint, downloading a released binary, verifying checksums — is future
// work once nova has actual binary releases.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update nova to the latest version (source-based)",
	Long: `update pulls the latest changes from the nova git repository and
rebuilds the binary. This command must be run from within the nova source
directory (a git clone).

Note: this is a source-based update mechanism. Binary release updates will be
supported in a future version once nova has published releases.`,
	Run: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	// Check we're in a git repo
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		red.Println("Error: 'nova update' must be run from the nova source directory.")
		fmt.Println("The current directory does not appear to be a git repository (.git not found).")
		os.Exit(1)
	}

	// Check that git is available
	if _, err := exec.LookPath("git"); err != nil {
		red.Println("Error: 'git' not found in PATH. Cannot update.")
		os.Exit(1)
	}

	// Check that go is available
	if _, err := exec.LookPath("go"); err != nil {
		red.Println("Error: 'go' not found in PATH. Cannot rebuild.")
		os.Exit(1)
	}

	// Step 1: git pull
	cyan.Println("========== Updating Nova ==========")
	cyan.Println("Pulling latest changes...")

	pullCmd := exec.Command("git", "pull")
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		red.Printf("Error: git pull failed: %v\n", err)
		os.Exit(1)
	}

	// Step 2: go build
	outputBin := "nova"
	if runtime.GOOS == "windows" {
		outputBin = "nova.exe"
	}

	cyan.Printf("Rebuilding binary (%s)...\n", outputBin)

	buildCmd := exec.Command("go", "build", "-o", outputBin, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		red.Printf("Error: go build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("-----------------------------------")
	green.Printf("✓ Nova updated successfully (%s)\n", outputBin)
	cyan.Println("===================================")
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
