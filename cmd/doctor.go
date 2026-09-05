/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	"errors"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Commands struct {
	Name        string   `json:"name"`
	Executables []string `json:"executables"`
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the health of your development environment",
	Long: `doctor inspects your machine for the tools GalactOS development relies on
	(Go, Docker, Git, and others as nova grows) and reports what's installed,
	what's missing, and what needs attention — a single command instead of
	checking each tool by hand.`,
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen)
		red := color.New(color.FgRed)
		yellow := color.New(color.FgYellow)
		cyan := color.New(color.FgCyan)

		cyan.Println("==========Running HealthChecks==========")
		test := []Commands{
			{Name: "GO", Executables: []string{"go", "version"}},
			{Name: "Docker", Executables: []string{"docker", "version"}},
			{Name: "Git", Executables: []string{"git", "--version"}},
			{Name: "Node.js", Executables: []string{"node", "--version"}},
			{Name: "npm", Executables: []string{"npm", "--version"}},
			{Name: "Python", Executables: []string{"python3", "--version"}},
			{Name: "pip", Executables: []string{"pip3", "--version"}},
		}

		for _, tool := range test {
			cyan.Printf("Checking %s...\n", tool.Name)

			comd := exec.Command(tool.Executables[0], tool.Executables[1:]...)
			err := comd.Run()

			switch {
			case err == nil:
				green.Printf("[SUCCESS] %v\n", tool.Name)

			case errors.Is(err, exec.ErrNotFound):
				red.Printf("[NOT FOUND] %v — not installed or not in PATH\n", tool.Name)

			case errors.As(err, new(*exec.ExitError)):
				yellow.Printf("[NOT RUNNING] %v — installed, but command failed (exit error)\n", tool.Name)

			default:
				red.Printf("[FAILED] %v — %v\n", tool.Name, err)
			}
		}
		cyan.Println("==========HealthChecks Completed==========")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// doctorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// doctorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
