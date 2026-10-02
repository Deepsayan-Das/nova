/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// execCmd represents nova container exec
var execCmd = &cobra.Command{
	Use:   "exec [container_id_or_name]",
	Short: "Execute a command inside a running container",
	Long: `exec opens an interactive shell or runs a command inside a running container
using 'podman exec'. By default, opens /bin/sh.

Examples:
  nova container exec my-container           — opens /bin/sh in the container
  nova container exec my-container -- ls -la — runs a specific command`,
	Run: func(cmd *cobra.Command, args []string) {
		var target string
		if len(args) > 0 {
			target = args[0]
		}

		if strings.TrimSpace(target) == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Container name or ID:",
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

		// Determine command to run
		execCommand := []string{"/bin/sh"}
		if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 && dashIdx < len(args) {
			execCommand = args[dashIdx:]
		}

		podmanArgs := []string{"exec", "-it", target}
		podmanArgs = append(podmanArgs, execCommand...)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Entering container '%s'...\n\n", target)
		_, err = RunPodmanCommand(podmanArgs, true)
		if err != nil {
			fmt.Printf("\nError: exec failed: %v\n", err)
			return
		}
	},
}

// rmCmd represents nova container rm
var rmCmd = &cobra.Command{
	Use:   "rm [container_id_or_name...]",
	Short: "Remove one or more stopped containers",
	Long: `rm removes stopped containers using 'podman rm'. Supports removing multiple
containers at once and force removal of running containers.

Examples:
  nova container rm my-container
  nova container rm -f my-container    — force remove (even if running)`,
	Run: func(cmd *cobra.Command, args []string) {
		forceFlag, _ := cmd.Flags().GetBool("force")

		targets := args
		if len(targets) == 0 {
			var target string
			err := survey.AskOne(&survey.Input{
				Message: "Container name(s) or ID(s) to remove (space-separated):",
			}, &target)
			if err != nil {
				fmt.Println(err)
				return
			}
			targets = strings.Fields(strings.TrimSpace(target))
		}

		if len(targets) == 0 {
			fmt.Println("Error: At least one container ID or name is required")
			return
		}

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, t := range targets {
			podmanArgs := []string{"rm"}
			if forceFlag {
				podmanArgs = append(podmanArgs, "-f")
			}
			podmanArgs = append(podmanArgs, t)

			out, err := RunPodmanCommand(podmanArgs, false)
			if err != nil {
				fmt.Printf("Error removing container '%s': %v\n", t, err)
				continue
			}
			fmt.Printf("✓ Removed container: %s\n", out)
		}
	},
}

// imagesCmd represents nova container images
var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "List local container images",
	Long:  `images lists all locally stored container images using 'podman images'.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		podmanArgs := []string{"images"}
		out, err := RunPodmanCommand(podmanArgs, false)
		if err != nil {
			fmt.Printf("Error listing images: %v\n", err)
			return
		}

		if out != "" {
			fmt.Println(out)
		} else {
			fmt.Println("No local images found.")
		}
	},
}

// rmiCmd represents nova container rmi (remove image)
var rmiCmd = &cobra.Command{
	Use:   "rmi [image_id_or_name...]",
	Short: "Remove one or more local container images",
	Long:  `rmi removes local container images using 'podman rmi'.`,
	Run: func(cmd *cobra.Command, args []string) {
		forceFlag, _ := cmd.Flags().GetBool("force")

		targets := args
		if len(targets) == 0 {
			var target string
			err := survey.AskOne(&survey.Input{
				Message: "Image name(s) or ID(s) to remove (space-separated):",
			}, &target)
			if err != nil {
				fmt.Println(err)
				return
			}
			targets = strings.Fields(strings.TrimSpace(target))
		}

		if len(targets) == 0 {
			fmt.Println("Error: At least one image ID or name is required")
			return
		}

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, t := range targets {
			podmanArgs := []string{"rmi"}
			if forceFlag {
				podmanArgs = append(podmanArgs, "-f")
			}
			podmanArgs = append(podmanArgs, t)

			out, err := RunPodmanCommand(podmanArgs, false)
			if err != nil {
				fmt.Printf("Error removing image '%s': %v\n", t, err)
				continue
			}
			fmt.Printf("✓ Removed image: %s\n", out)
		}
	},
}

// inspectContainerCmd represents nova container inspect
var inspectContainerCmd = &cobra.Command{
	Use:   "inspect [container_id_or_name]",
	Short: "Display detailed information about a container",
	Long:  `inspect shows low-level information about a container using 'podman inspect'.`,
	Run: func(cmd *cobra.Command, args []string) {
		var target string
		if len(args) > 0 {
			target = args[0]
		}

		if strings.TrimSpace(target) == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Container name or ID:",
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

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		podmanArgs := []string{"inspect", target}
		out, err := RunPodmanCommand(podmanArgs, false)
		if err != nil {
			fmt.Printf("Error inspecting container '%s': %v\n", target, err)
			return
		}
		fmt.Println(out)
	},
}

func init() {
	rmCmd.Flags().BoolP("force", "f", false, "Force remove (even if running)")
	rmiCmd.Flags().BoolP("force", "f", false, "Force remove image")

	containerCmd.AddCommand(execCmd)
	containerCmd.AddCommand(rmCmd)
	containerCmd.AddCommand(imagesCmd)
	containerCmd.AddCommand(rmiCmd)
	containerCmd.AddCommand(inspectContainerCmd)
}
