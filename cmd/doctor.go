/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Deepsayan-Das/nova/Types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed doctor.yaml
var defaultDoctorYAML []byte

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor [tool]",
	Short: "Check the health of your development environment",
	Long: `doctor inspects your machine for the tools GalactOS development relies on
(Go, Docker, Git, and others as nova grows) using configurations from doctor.yaml.
	
You can test all tools by running:
  nova doctor

Or test a specific tool (e.g. Go) by passing its name:
  nova doctor go`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var yamlData []byte
		if data, err := os.ReadFile("doctor.yaml"); err == nil {
			yamlData = data
		} else {
			yamlData = defaultDoctorYAML
		}
		var config Types.DoctorConfig
		_ = yaml.Unmarshal(yamlData, &config)

		var completions []string
		for _, check := range config.Checks {
			if check.Key != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", check.Key, check.Description))
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		green := color.New(color.FgGreen)
		red := color.New(color.FgRed)
		yellow := color.New(color.FgYellow)
		cyan := color.New(color.FgCyan)

		cyan.Println("==========Running HealthChecks==========")

		var yamlData []byte

		// Try reading local doctor.yaml first if present
		if data, readErr := os.ReadFile("doctor.yaml"); readErr == nil {
			yamlData = data
		} else {
			yamlData = defaultDoctorYAML
		}

		var config Types.DoctorConfig
		if err := yaml.Unmarshal(yamlData, &config); err != nil {
			red.Printf("Error parsing doctor.yaml: %v\n", err)
			return
		}

		checksToRun := config.Checks

		// Filter by target tool if positional arguments were passed (e.g. nova doctor go)
		if len(args) > 0 {
			var filtered []Types.ToolCheck
			for _, check := range config.Checks {
				for _, arg := range args {
					if strings.EqualFold(check.Key, arg) || strings.EqualFold(check.Name, arg) {
						filtered = append(filtered, check)
						break
					}
				}
			}

			if len(filtered) == 0 {
				var available []string
				for _, check := range config.Checks {
					if check.Key != "" {
						available = append(available, fmt.Sprintf("%s (%s)", check.Name, check.Key))
					} else {
						available = append(available, check.Name)
					}
				}
				yellow.Printf("No matching health checks found for target(s): %s\n", strings.Join(args, ", "))
				cyan.Printf("Available tools: %s\n", strings.Join(available, ", "))
				cyan.Println("==========HealthChecks Completed==========")
				return
			}

			checksToRun = filtered
		}

		for _, tool := range checksToRun {
			cyan.Printf("\nTarget: %s\n", tool.Name)
			if tool.Description != "" {
				fmt.Printf("Description: %s\n", tool.Description)
			}

			// Elaborate tests
			if len(tool.Tests) > 0 {
				for _, t := range tool.Tests {
					if len(t.Command) == 0 {
						continue
					}
					cyan.Printf("  Checking %s [%s]...\n", t.Name, strings.Join(t.Command, " "))
					comd := exec.Command(t.Command[0], t.Command[1:]...)
					runErr := comd.Run()

					switch {
					case runErr == nil:
						green.Printf("  [SUCCESS] %s - %s\n", tool.Name, t.Name)
					case errors.Is(runErr, exec.ErrNotFound):
						red.Printf("  [NOT FOUND] %s - %s (executable '%s' not in PATH)\n", tool.Name, t.Name, t.Command[0])
					case errors.As(runErr, new(*exec.ExitError)):
						yellow.Printf("  [NOT RUNNING] %s - %s (command failed)\n", tool.Name, t.Name)
					default:
						red.Printf("  [FAILED] %s - %s (%v)\n", tool.Name, t.Name, runErr)
					}
				}
			} else if len(tool.Executables) > 0 {
				// Legacy simple executables check
				cyan.Printf("  Checking %s...\n", tool.Name)
				comd := exec.Command(tool.Executables[0], tool.Executables[1:]...)
				runErr := comd.Run()

				switch {
				case runErr == nil:
					green.Printf("  [SUCCESS] %s\n", tool.Name)
				case errors.Is(runErr, exec.ErrNotFound):
					red.Printf("  [NOT FOUND] %s — not installed or not in PATH\n", tool.Name)
				case errors.As(runErr, new(*exec.ExitError)):
					yellow.Printf("  [NOT RUNNING] %s — command failed\n", tool.Name)
				default:
					red.Printf("  [FAILED] %s — %v\n", tool.Name, runErr)
				}
			}
		}
		cyan.Println("\n==========HealthChecks Completed==========")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
