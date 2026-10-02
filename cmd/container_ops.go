/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	psAll      bool
	logsFollow bool
)

// psCmd represents nova container ps
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers using Podman",
	Long:  `Displays running containers by executing 'podman ps'. Use -a to show all containers including stopped ones.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := ContainerPsConfig{All: psAll}
		podmanArgs := BuildPodmanPsArgs(cfg)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		out, err := RunPodmanCommand(podmanArgs, false)
		if err != nil {
			fmt.Printf("Error listing containers: %v\n", err)
			return
		}

		if out != "" {
			fmt.Println(out)
		} else {
			fmt.Println("No containers found.")
		}
	},
}

// stopCmd represents nova container stop
var stopCmd = &cobra.Command{
	Use:   "stop [container_id_or_name]",
	Short: "Stop a running container using Podman",
	Long:  `Stops one or more running containers using 'podman stop'.`,
	Run: func(cmd *cobra.Command, args []string) {
		var target string
		if len(args) > 0 {
			target = args[0]
		}

		if strings.TrimSpace(target) == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Container name or ID to stop:",
			}, &target)
			if err != nil {
				fmt.Println(err)
				return
			}
			target = strings.TrimSpace(target)
		}

		if target == "" {
			fmt.Println("Error: Container ID or name is required")
			return
		}

		podmanArgs := BuildPodmanStopArgs(target)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		out, err := RunPodmanCommand(podmanArgs, false)
		if err != nil {
			fmt.Printf("Error stopping container '%s': %v\n", target, err)
			return
		}

		fmt.Printf("✓ Stopped container: %s\n", out)
	},
}

// logsCmd represents nova container logs
var logsCmd = &cobra.Command{
	Use:   "logs [container_id_or_name]",
	Short: "Fetch logs of a container using Podman",
	Long:  `Fetches and streams logs from a container using 'podman logs'. Supports -f/--follow flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		var target string
		if len(args) > 0 {
			target = args[0]
		}

		if strings.TrimSpace(target) == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Container name or ID to view logs:",
			}, &target)
			if err != nil {
				fmt.Println(err)
				return
			}
			target = strings.TrimSpace(target)
		}

		if target == "" {
			fmt.Println("Error: Container ID or name is required")
			return
		}

		cfg := ContainerLogsConfig{
			ContainerID: target,
			Follow:      logsFollow,
		}

		podmanArgs := BuildPodmanLogsArgs(cfg)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		if logsFollow {
			_, err = RunPodmanCommand(podmanArgs, true)
			if err != nil {
				fmt.Printf("\nError reading container logs: %v\n", err)
			}
		} else {
			out, err := RunPodmanCommand(podmanArgs, false)
			if err != nil {
				fmt.Printf("Error reading container logs: %v\n", err)
				return
			}
			if out != "" {
				fmt.Println(out)
			}
		}
	},
}

func init() {
	psCmd.Flags().BoolVarP(&psAll, "all", "a", false, "Show all containers (default shows just running)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output in real time")

	containerCmd.AddCommand(psCmd)
	containerCmd.AddCommand(stopCmd)
	containerCmd.AddCommand(logsCmd)
}
