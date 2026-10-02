/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	buildFile string
	buildTag  string
)

// buildCmd represents the nova container build command
var buildCmd = &cobra.Command{
	Use:   "build [context_path]",
	Short: "Build a container image using Podman",
	Long: `Builds a container image using Podman from a Dockerfile in the current directory or specified path.
Real-time build logs are streamed directly to the terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		contextPath := "."
		if len(args) > 0 {
			contextPath = args[0]
		}

		if buildFile == "" {
			buildFile = "Dockerfile"
		}

		// Verify Dockerfile exists
		dockerfilePath := buildFile
		if !filepath.IsAbs(dockerfilePath) && contextPath != "." {
			dockerfilePath = filepath.Join(contextPath, buildFile)
		}
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			fmt.Printf("Error: Dockerfile not found at path '%s'\n", dockerfilePath)
			return
		}

		// Prompt for tag if not provided
		if strings.TrimSpace(buildTag) == "" {
			var tagAnswer string
			err := survey.AskOne(&survey.Input{
				Message: "Image tag/name (e.g. my-app:latest):",
				Default: "my-app:latest",
			}, &tagAnswer)
			if err != nil {
				fmt.Println(err)
				return
			}
			buildTag = strings.TrimSpace(tagAnswer)
		}

		cfg := ContainerBuildConfig{
			Dockerfile:  buildFile,
			Tag:         buildTag,
			ContextPath: contextPath,
		}

		podmanArgs := BuildPodmanBuildArgs(cfg)

		fmt.Printf("Building container image '%s' using Podman...\n\n", buildTag)

		err := CheckPodmanInstalled()
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = RunPodmanCommand(podmanArgs, true)
		if err != nil {
			fmt.Printf("\nError: Podman build failed: %v\n", err)
			return
		}

		fmt.Printf("\n✓ Successfully built image '%s'\n", buildTag)
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildFile, "file", "f", "Dockerfile", "Path to the Dockerfile")
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "Name and optionally a tag in the 'name:tag' format")
	containerCmd.AddCommand(buildCmd)
}
