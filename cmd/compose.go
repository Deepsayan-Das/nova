/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	Image       string            `yaml:"image,omitempty"`
	Build       string            `yaml:"build,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
}

type ComposeConfig struct {
	Version  string                   `yaml:"version"`
	Services map[string]ServiceConfig `yaml:"services"`
}

// GenerateCompose generates valid docker-compose YAML from ComposeConfig
func GenerateCompose(cfg ComposeConfig) string {
	if cfg.Version == "" {
		cfg.Version = "3.8"
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return ""
	}
	return string(out)
}

// composeCmd represents the compose command
var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Generate a docker-compose.yml file interactively",
	Long:  `Compose interactively creates a docker-compose.yml file with services, ports, environment variables, and dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		var countStr string
		err := survey.AskOne(&survey.Input{
			Message: "How many services to configure?",
			Default: "1",
		}, &countStr)
		if err != nil {
			fmt.Println(err)
			return
		}

		numServices, err := strconv.Atoi(strings.TrimSpace(countStr))
		if err != nil || numServices < 1 {
			numServices = 1
		}

		cfg := ComposeConfig{
			Version:  "3.8",
			Services: make(map[string]ServiceConfig),
		}

		var createdNames []string

		for i := 0; i < numServices; i++ {
			fmt.Printf("\n--- Service %d of %d ---\n", i+1, numServices)

			var name string
			defaultName := fmt.Sprintf("service%d", i+1)
			if i == 0 {
				defaultName = "web"
			}
			err := survey.AskOne(&survey.Input{
				Message: "Service name:",
				Default: defaultName,
			}, &name)
			if err != nil {
				fmt.Println(err)
				return
			}
			name = strings.TrimSpace(name)

			var sourceChoice string
			err = survey.AskOne(&survey.Select{
				Message: "Source type for service:",
				Options: []string{"Base Image", "Local Build Context (.)"},
				Default: "Base Image",
			}, &sourceChoice)
			if err != nil {
				fmt.Println(err)
				return
			}

			svc := ServiceConfig{}

			if sourceChoice == "Local Build Context (.)" {
				svc.Build = "."
			} else {
				var image string
				err := survey.AskOne(&survey.Input{
					Message: "Base image (e.g. node:20-alpine, postgres:15):",
				}, &image)
				if err != nil {
					fmt.Println(err)
					return
				}
				svc.Image = strings.TrimSpace(image)
			}

			var portsStr string
			err = survey.AskOne(&survey.Input{
				Message: "Port mappings (comma-separated, e.g. 3000:3000, leave blank if none):",
			}, &portsStr)
			if err != nil {
				fmt.Println(err)
				return
			}
			portsStr = strings.TrimSpace(portsStr)
			if portsStr != "" {
				rawPorts := strings.Split(portsStr, ",")
				for _, p := range rawPorts {
					p = strings.TrimSpace(p)
					if p != "" {
						svc.Ports = append(svc.Ports, p)
					}
				}
			}

			var envCountStr string
			err = survey.AskOne(&survey.Input{
				Message: "How many environment variables for this service?",
				Default: "0",
			}, &envCountStr)
			if err != nil {
				fmt.Println(err)
				return
			}
			envCount, _ := strconv.Atoi(strings.TrimSpace(envCountStr))
			if envCount > 0 {
				svc.Environment = make(map[string]string)
				for k := 0; k < envCount; k++ {
					var kvStr string
					err := survey.AskOne(&survey.Input{
						Message: fmt.Sprintf("Environment variable #%d (KEY=VALUE):", k+1),
					}, &kvStr)
					if err != nil {
						fmt.Println(err)
						return
					}
					kvStr = strings.TrimSpace(kvStr)
					if kvStr != "" {
						parts := strings.SplitN(kvStr, "=", 2)
						if len(parts) == 2 {
							svc.Environment[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
						} else {
							svc.Environment[parts[0]] = ""
						}
					}
				}
			}

			if len(createdNames) > 0 {
				var deps []string
				err := survey.AskOne(&survey.MultiSelect{
					Message: "Select dependencies (depends_on):",
					Options: createdNames,
				}, &deps)
				if err == nil && len(deps) > 0 {
					svc.DependsOn = deps
				}
			}

			cfg.Services[name] = svc
			createdNames = append(createdNames, name)
		}

		composeYaml := GenerateCompose(cfg)
		err = os.WriteFile("docker-compose.yml", []byte(composeYaml), 0644)
		if err != nil {
			fmt.Printf("Error writing docker-compose.yml: %v\n", err)
			return
		}

		fmt.Println("✓ docker-compose.yml generated at ./docker-compose.yml")
	},
}

func init() {
	containerCmd.AddCommand(composeCmd)
}
