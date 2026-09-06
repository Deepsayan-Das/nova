/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Scaffold a new project",
	Long: `init creates a new project directory with a starter file structure
based on the chosen language and template (e.g. --lang=go --template=cli).

If a path or directory name is passed as an argument (e.g. ./ or my-app), scaffolding
will happen in that location. If no path is provided, you will be prompted for a project name
which will be created as a directory in the current location.`,
	Run: runInit,
}

func runInit(cmd *cobra.Command, args []string) {
	lang, _ := cmd.Flags().GetString("lang")
	templateName, _ := cmd.Flags().GetString("template")
	gitFlag, _ := cmd.Flags().GetBool("git")

	allTemplates := GetAllTemplates()

	if lang == "" || templateName == "" {
		fmt.Println("Please specify both --lang and --template flags.")
		fmt.Println("\nAvailable templates:")
		for _, t := range allTemplates {
			fmt.Printf("  - Language: %-10s Template: %s\n", t.Language, t.Name)
		}
		os.Exit(1)
	}

	selected := GetTemplate(lang, templateName)

	if selected == nil {
		fmt.Printf("No template found for language '%s' and template '%s'.\n", lang, templateName)
		fmt.Println("\nAvailable templates:")
		for _, t := range allTemplates {
			fmt.Printf("  - Language: %-10s Template: %s\n", t.Language, t.Name)
		}
		os.Exit(1)
	}

	var targetDir string
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		targetDir = strings.TrimSpace(args[0])
	} else {
		input, err := promptUser("Enter project name: ")
		if err != nil {
			fmt.Printf("Error reading project name: %v\n", err)
			os.Exit(1)
		}
		targetDir = input
		if targetDir == "" {
			fmt.Println("Project name cannot be empty.")
			os.Exit(1)
		}
	}

	nonEmpty, err := isDirNonEmpty(targetDir)
	if err != nil {
		fmt.Printf("Error checking target directory '%s': %v\n", targetDir, err)
		os.Exit(1)
	}

	if nonEmpty {
		confirm, err := promptUser(fmt.Sprintf("Target directory '%s' already exists and is not empty. Overwrite existing files? (y/N): ", targetDir))
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			os.Exit(1)
		}
		confirm = strings.ToLower(strings.TrimSpace(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Initialization cancelled.")
			os.Exit(1)
		}
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("Error creating project target directory '%s': %v\n", targetDir, err)
		os.Exit(1)
	}

	fmt.Printf("Initializing project in '%s' with Language: %s, Template: %s...\n", targetDir, selected.Language, selected.Name)

	for _, dir := range selected.Dirs {
		dirPath := filepath.Join(targetDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			os.Exit(1)
		}
		fmt.Printf("  Created directory: %s\n", dirPath)
	}

	for _, file := range selected.Files {
		filePath := filepath.Join(targetDir, file.Path)
		if dir := filepath.Dir(filePath); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}
		if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			os.Exit(1)
		}
		fmt.Printf("  Created file: %s\n", filePath)
	}

	if gitFlag {
		gitDir := filepath.Join(targetDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			gitCmd := exec.Command("git", "init")
			gitCmd.Dir = targetDir
			if err := gitCmd.Run(); err != nil {
				fmt.Printf("  Warning: failed to initialize Git repository: %v\n", err)
			} else {
				fmt.Println("  Initialized empty Git repository.")
			}
		}
	}

	metadataContent := fmt.Sprintf("language: %s\ntemplate: %s\ncreated: %s\nnova_version: %s\n",
		selected.Language,
		selected.Name,
		time.Now().Format("2006-01-02"),
		Version,
	)
	metadataPath := filepath.Join(targetDir, ".nova.yaml")
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		fmt.Printf("Error creating metadata file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Created metadata file: %s\n", metadataPath)

	fmt.Println("Project initialization complete!")
}

func promptUser(prompt string) (string, error) {
	if prompt != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func isDirNonEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
}

func init() {
	projectCmd.AddCommand(initCmd)

	initCmd.Flags().String("lang", "", "Language for the new project (e.g. go, js)")
	initCmd.Flags().String("template", "", "Template to use (e.g. cli, api, mvc, microservices)")
	initCmd.Flags().Bool("git", true, "Initialize a Git repository (default true)")
}
