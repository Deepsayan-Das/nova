/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// dockerignoreCmd generates a .dockerignore file
var dockerignoreCmd = &cobra.Command{
	Use:   "dockerignore",
	Short: "Generate a .dockerignore file",
	Long: `dockerignore creates a .dockerignore file tailored to your project language.
This prevents unnecessary files (node_modules, .git, build artifacts, etc.)
from being sent to the Docker build context, reducing image size and build time.`,
	Run: func(cmd *cobra.Command, args []string) {
		var lang string
		var addCustom bool

		survey.AskOne(&survey.Select{
			Message: "Project language/stack:",
			Options: []string{"node", "go", "python", "rust", "java", "generic"},
			Default: "generic",
		}, &lang)

		content := generateDockerignore(lang)

		survey.AskOne(&survey.Confirm{Message: "Add custom ignore patterns?", Default: false}, &addCustom)
		if addCustom {
			var extra string
			survey.AskOne(&survey.Input{
				Message: "Additional patterns (comma-separated):",
			}, &extra)
			if strings.TrimSpace(extra) != "" {
				content += "\n# Custom patterns\n"
				for _, pattern := range strings.Split(extra, ",") {
					p := strings.TrimSpace(pattern)
					if p != "" {
						content += p + "\n"
					}
				}
			}
		}

		if err := os.WriteFile(".dockerignore", []byte(content), 0644); err != nil {
			fmt.Printf("Error writing .dockerignore: %v\n", err)
			return
		}
		fmt.Println("✓ Generated .dockerignore")
	},
}

func generateDockerignore(lang string) string {
	common := `# Version control
.git
.gitignore
.gitattributes

# IDE and editor files
.vscode
.idea
*.swp
*.swo
*~
.DS_Store
Thumbs.db

# CI/CD
.github
.gitlab-ci.yml
Jenkinsfile

# Docker
Dockerfile
docker-compose*.yml
.dockerignore

# Documentation
README.md
LICENSE
CHANGELOG.md
docs/

# Nova metadata
.nova.yaml
`

	switch lang {
	case "node":
		return common + `
# Node.js specific
node_modules
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.npm
.yarn
.pnpm-store

# Build output
dist
build
out
.next
.nuxt

# Test & coverage
coverage
.nyc_output

# Environment
.env
.env.local
.env.*.local
`
	case "go":
		return common + `
# Go specific
vendor/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
bin/
dist/

# Go coverage
coverage.out
coverage.html

# Environment
.env
`
	case "python":
		return common + `
# Python specific
__pycache__
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
*.egg

# Virtual environments
.venv
venv
env
ENV

# Test & coverage
htmlcov/
.coverage
.coverage.*
.pytest_cache

# Environment
.env
.env.local
`
	case "rust":
		return common + `
# Rust specific
target/
**/*.rs.bk
Cargo.lock

# Environment
.env
`
	case "java":
		return common + `
# Java specific
*.class
*.jar
*.war
*.ear
*.nar
hs_err_pid*

# Build output
target/
build/
out/
.gradle/

# Environment
.env
`
	default:
		return common + `
# Environment
.env
.env.local

# Build output
build/
dist/
out/
bin/

# Test & coverage
coverage/
`
	}
}

func init() {
	containerCmd.AddCommand(dockerignoreCmd)
}
