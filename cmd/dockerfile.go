/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type DockerfileConfig struct {
	BaseImage  string `survey:"base_image"`
	Version    string `survey:"version"`
	Port       string `survey:"port"`
	WorkDir    string `survey:"work_dir"`
	EntryCmd   string `survey:"entry_cmd"`
	Multistage bool   `survey:"multistage"`
}

func formatCmd(entry string) string {
	parts := strings.Fields(entry)
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = fmt.Sprintf("%q", p)
	}
	return strings.Join(quoted, ", ")
}

func GenerateDockerfile(cfg DockerfileConfig) string {
	var sb strings.Builder

	if cfg.Multistage {
		switch cfg.BaseImage {
		case "go":
			version := cfg.Version
			if version == "" {
				version = "1.22-alpine"
			}
			sb.WriteString(fmt.Sprintf("# Build stage\nFROM golang:%s AS builder\n", version))
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString("COPY go.mod go.sum ./\n")
			sb.WriteString("RUN go mod download\n")
			sb.WriteString("COPY . .\n")
			sb.WriteString("RUN CGO_ENABLED=0 GOOS=linux go build -o app .\n\n")

			sb.WriteString("# Runtime stage\nFROM alpine:latest\n")
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString(fmt.Sprintf("COPY --from=builder %s/app ./app\n", cfg.WorkDir))
			if cfg.Port != "" {
				sb.WriteString(fmt.Sprintf("EXPOSE %s\n", cfg.Port))
			}
			if cfg.EntryCmd != "" {
				sb.WriteString(fmt.Sprintf("CMD [%s]\n", formatCmd(cfg.EntryCmd)))
			} else {
				sb.WriteString("CMD [\"./app\"]\n")
			}
			return sb.String()

		case "node":
			version := cfg.Version
			if version == "" {
				version = "20"
			}
			sb.WriteString(fmt.Sprintf("# Build stage\nFROM node:%s AS builder\n", version))
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString("COPY package*.json ./\n")
			sb.WriteString("RUN npm install\n")
			sb.WriteString("COPY . .\n")
			sb.WriteString("RUN npm run build\n\n")

			sb.WriteString(fmt.Sprintf("# Runtime stage\nFROM node:%s-alpine\n", version))
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString("COPY package*.json ./\n")
			sb.WriteString("RUN npm install --only=production\n")
			sb.WriteString(fmt.Sprintf("COPY --from=builder %s/dist ./dist\n", cfg.WorkDir))
			if cfg.Port != "" {
				sb.WriteString(fmt.Sprintf("EXPOSE %s\n", cfg.Port))
			}
			if cfg.EntryCmd != "" {
				sb.WriteString(fmt.Sprintf("CMD [%s]\n", formatCmd(cfg.EntryCmd)))
			} else {
				sb.WriteString("CMD [\"npm\", \"start\"]\n")
			}
			return sb.String()

		case "python":
			sb.WriteString("# Note: Python projects typically use single-stage builds as they are interpreted.\n")
			// Fall back to single stage below

		default:
			sb.WriteString(fmt.Sprintf("# Build stage\nFROM %s:%s AS builder\n", cfg.BaseImage, cfg.Version))
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString("COPY . .\n")
			sb.WriteString("RUN echo \"Building application...\"\n\n")

			sb.WriteString("# Runtime stage\nFROM alpine:latest\n")
			sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
			sb.WriteString(fmt.Sprintf("COPY --from=builder %s ./\n", cfg.WorkDir))
			if cfg.Port != "" {
				sb.WriteString(fmt.Sprintf("EXPOSE %s\n", cfg.Port))
			}
			if cfg.EntryCmd != "" {
				sb.WriteString(fmt.Sprintf("CMD [%s]\n", formatCmd(cfg.EntryCmd)))
			}
			return sb.String()
		}
	}

	// Single-stage build generation
	sb.WriteString(fmt.Sprintf("FROM %s:%s\n", cfg.BaseImage, cfg.Version))
	sb.WriteString(fmt.Sprintf("WORKDIR %s\n", cfg.WorkDir))
	switch cfg.BaseImage {
	case "node":
		sb.WriteString("COPY package*.json ./\n")
		sb.WriteString("RUN npm install\n\n")
	case "python":
		sb.WriteString("COPY requirements.txt ./\n")
		sb.WriteString("RUN pip install -r requirements.txt\n\n")
	case "go":
		sb.WriteString("COPY go.mod go.sum ./\n")
		sb.WriteString("RUN go mod download\n\n")
	}
	sb.WriteString("COPY . .\n\n")
	if cfg.Port != "" {
		sb.WriteString(fmt.Sprintf("EXPOSE %s\n\n", cfg.Port))
	}
	if cfg.EntryCmd != "" {
		sb.WriteString(fmt.Sprintf("CMD [%s]\n", formatCmd(cfg.EntryCmd)))
	}

	return sb.String()
}

// dockerfileCmd represents the dockerfile command
var dockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var qs = []*survey.Question{
			{
				Name: "base_image",
				Prompt: &survey.Select{
					Message: "Choose a base image:",
					Options: []string{"node", "python", "go", "custom"},
				},
			},
			{
				Name: "version",
				Prompt: &survey.Input{
					Message: "Version/tag (e.g. 20-alpine, 3.12-slim):",
				},
			},
			{
				Name: "port",
				Prompt: &survey.Input{
					Message: "Port to expose:",
					Default: "3000",
				},
			},
			{
				Name: "work_dir",
				Prompt: &survey.Input{
					Message: "Working directory:",
					Default: "/app",
				},
			},
			{
				Name: "entry_cmd",
				Prompt: &survey.Input{
					Message: "Entry command (e.g. npm start, python app.py):",
				},
			},
			{
				Name: "multistage",
				Prompt: &survey.Confirm{
					Message: "Does this project need a build step (compiled/bundled)?",
					Default: false,
				},
			},
		}
		answers := DockerfileConfig{}
		err := survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err)
			return
		}

		dockerfile := GenerateDockerfile(answers)
		err = os.WriteFile("Dockerfile", []byte(dockerfile), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("✓ Dockerfile generated at ./Dockerfile")
	},
}

func init() {
	containerCmd.AddCommand(dockerfileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dockerfileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dockerfileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
