/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	runPort   string
	runEnv    []string
	runDetach bool
	runName   string
)

// runCmd represents the nova container run command
var runCmd = &cobra.Command{
	Use:   "run [image]",
	Short: "Run a container from an image using Podman",
	Long: `Runs a container using Podman with options for port mappings, environment variables, detached mode, and container names.
If required options are omitted, nova will interactively prompt for them.`,
	Run: func(cmd *cobra.Command, args []string) {
		var image string
		if len(args) > 0 {
			image = args[0]
		}

		if strings.TrimSpace(image) == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Image name to run (e.g. node:20-alpine, my-app:latest):",
			}, &image)
			if err != nil {
				fmt.Println(err)
				return
			}
			image = strings.TrimSpace(image)
		}

		if image == "" {
			fmt.Println("Error: Image name is required to run container")
			return
		}

		cfg := ContainerRunConfig{
			Image:  image,
			Detach: runDetach,
			Name:   strings.TrimSpace(runName),
		}

		if runPort != "" {
			cfg.Ports = append(cfg.Ports, runPort)
		}
		if len(runEnv) > 0 {
			cfg.Env = append(cfg.Env, runEnv...)
		}

		// Interactive prompts if flags were not provided
		if !cmd.Flags().Changed("port") && len(cfg.Ports) == 0 {
			var portAnswer string
			err := survey.AskOne(&survey.Input{
				Message: "Port mapping (host:container, e.g. 8080:80, leave blank to skip):",
			}, &portAnswer)
			if err == nil && strings.TrimSpace(portAnswer) != "" {
				cfg.Ports = append(cfg.Ports, strings.TrimSpace(portAnswer))
			}
		}

		if !cmd.Flags().Changed("env") && len(cfg.Env) == 0 {
			var envCountStr string
			err := survey.AskOne(&survey.Input{
				Message: "How many environment variables to pass?",
				Default: "0",
			}, &envCountStr)
			if err == nil {
				envCount, _ := strconv.Atoi(strings.TrimSpace(envCountStr))
				for i := 0; i < envCount; i++ {
					var kvStr string
					err := survey.AskOne(&survey.Input{
						Message: fmt.Sprintf("Environment variable #%d (KEY=VALUE):", i+1),
					}, &kvStr)
					if err == nil && strings.TrimSpace(kvStr) != "" {
						cfg.Env = append(cfg.Env, strings.TrimSpace(kvStr))
					}
				}
			}
		}

		if !cmd.Flags().Changed("detach") {
			var detachConfirm bool
			err := survey.AskOne(&survey.Confirm{
				Message: "Run container in detached mode (-d)?",
				Default: false,
			}, &detachConfirm)
			if err == nil {
				cfg.Detach = detachConfirm
			}
		}

		if !cmd.Flags().Changed("name") && cfg.Name == "" {
			var nameAnswer string
			err := survey.AskOne(&survey.Input{
				Message: "Container name (optional, leave blank for auto-generated name):",
			}, &nameAnswer)
			if err == nil {
				cfg.Name = strings.TrimSpace(nameAnswer)
			}
		}

		podmanArgs := BuildPodmanRunArgs(cfg)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		if cfg.Detach {
			containerID, err := RunPodmanCommand(podmanArgs, false)
			if err != nil {
				fmt.Printf("Error running container in detached mode: %v\n", err)
				return
			}
			fmt.Printf("✓ Container started in detached mode.\nID: %s\n", containerID)
		} else {
			fmt.Printf("Starting container '%s' in foreground mode...\n\n", image)
			_, err := RunPodmanCommand(podmanArgs, true)
			if err != nil {
				fmt.Printf("\nError: Container execution failed: %v\n", err)
				return
			}
		}
	},
}

func init() {
	runCmd.Flags().StringVarP(&runPort, "port", "p", "", "Port mapping in host:container format (e.g. 8080:80)")
	runCmd.Flags().StringSliceVarP(&runEnv, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "Run container in background (detached mode)")
	runCmd.Flags().StringVar(&runName, "name", "", "Assign a name to the container")

	containerCmd.AddCommand(runCmd)
}
