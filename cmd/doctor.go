/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the health of your development environment",
	Long: `doctor inspects your machine for the tools GalactOS development relies on
	(Go, Docker, Git, and others as nova grows) and reports what's installed,
	what's missing, and what needs attention — a single command instead of
	checking each tool by hand.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("==========Running HealthChecks==========")

		goCmd := exec.Command("go", "version")
		err := goCmd.Run()
		if err != nil {
			fmt.Println("[FAILED] : GO")

		} else {
			fmt.Println("[PASSED] : GO")
		}
		dockerCmd := exec.Command("docker", "version")
		err = dockerCmd.Run()
		if err != nil {
			fmt.Println("[FAILED] : Docker")

		} else {
			fmt.Println("[PASSED] : Docker")
		}
		gitCmd := exec.Command("git", "--version")
		err = gitCmd.Run()
		if err != nil {
			fmt.Println("[FAILED] : Git")

		} else {
			fmt.Println("[PASSED] : Git")
		}
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
