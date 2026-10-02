/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
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

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Generate CI/CD pipeline configuration",
	Long: `ci generates CI/CD pipeline configuration files for your project.

Supported providers:
  - GitHub Actions (.github/workflows/ci.yml)
  - GitLab CI (.gitlab-ci.yml)
  - Jenkinsfile (Jenkinsfile)

The generated pipeline includes build, test, lint, and optional Docker build/push stages.`,
	Run: runCI,
}

func runCI(cmd *cobra.Command, args []string) {
	var provider, lang string
	var addDocker, addDeploy bool

	survey.AskOne(&survey.Select{
		Message: "CI/CD provider:",
		Options: []string{"GitHub Actions", "GitLab CI", "Jenkinsfile"},
		Default: "GitHub Actions",
	}, &provider)

	survey.AskOne(&survey.Select{
		Message: "Project language:",
		Options: []string{"go", "node", "python", "rust", "java", "generic"},
		Default: "go",
	}, &lang)

	survey.AskOne(&survey.Confirm{Message: "Add Docker build & push stage?", Default: true}, &addDocker)
	survey.AskOne(&survey.Confirm{Message: "Add deploy stage (placeholder)?", Default: false}, &addDeploy)

	switch provider {
	case "GitHub Actions":
		generateGitHubActions(lang, addDocker, addDeploy)
	case "GitLab CI":
		generateGitLabCI(lang, addDocker, addDeploy)
	case "Jenkinsfile":
		generateJenkinsfile(lang, addDocker, addDeploy)
	}
}

func generateGitHubActions(lang string, addDocker, addDeploy bool) {
	dir := filepath.Join(".github", "workflows")
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	var buildSteps, testSteps, lintSteps string

	switch lang {
	case "go":
		buildSteps = `      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go build -v ./...`
		testSteps = `      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.out`
		lintSteps = `      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest`
	case "node":
		buildSteps = `      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm run build --if-present`
		testSteps = `      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npm test`
		lintSteps = `      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
      - run: npm ci
      - run: npx eslint . --ext .js,.jsx,.ts,.tsx`
	case "python":
		buildSteps = `      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install -r requirements.txt`
		testSteps = `      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install -r requirements.txt
      - run: python -m pytest --cov=. --cov-report=xml`
		lintSteps = `      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install flake8 black
      - run: flake8 .
      - run: black --check .`
	case "rust":
		buildSteps = `      - uses: dtolnay/rust-toolchain@stable
      - run: cargo build --verbose`
		testSteps = `      - uses: dtolnay/rust-toolchain@stable
      - run: cargo test --verbose`
		lintSteps = `      - uses: dtolnay/rust-toolchain@stable
        with:
          components: clippy, rustfmt
      - run: cargo clippy -- -D warnings
      - run: cargo fmt --check`
	case "java":
		buildSteps = `      - uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'
      - run: ./gradlew build`
		testSteps = `      - uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'
      - run: ./gradlew test`
		lintSteps = `      - uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'
      - run: ./gradlew check`
	default:
		buildSteps = `      - run: echo "Add your build steps here"`
		testSteps = `      - run: echo "Add your test steps here"`
		lintSteps = `      - run: echo "Add your lint steps here"`
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`name: CI

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
%s

  build:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4
%s

  test:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4
%s
`, lintSteps, buildSteps, testSteps))

	if addDocker {
		sb.WriteString(`
  docker:
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ghcr.io/${{ github.repository }}:latest
            ghcr.io/${{ github.repository }}:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
`)
	}

	if addDeploy {
		sb.WriteString(`
  deploy:
    runs-on: ubuntu-latest
    needs: [test` + func() string {
			if addDocker {
				return ", docker"
			}
			return ""
		}() + `]
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        run: |
          echo "Add your deployment steps here"
          echo "Examples: kubectl apply, helm upgrade, SSH deploy, etc."
`)
	}

	outPath := filepath.Join(dir, "ci.yml")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		fmt.Printf("Error writing CI file: %v\n", err)
		return
	}
	fmt.Printf("✓ Generated GitHub Actions workflow at %s\n", filepath.ToSlash(outPath))
}

func generateGitLabCI(lang string, addDocker, addDeploy bool) {
	var buildScript, testScript, lintScript string

	switch lang {
	case "go":
		buildScript = "go build -v ./..."
		testScript = "go test -v -race -coverprofile=coverage.out ./..."
		lintScript = "golangci-lint run"
	case "node":
		buildScript = "npm ci && npm run build --if-present"
		testScript = "npm ci && npm test"
		lintScript = "npm ci && npx eslint ."
	case "python":
		buildScript = "pip install -r requirements.txt"
		testScript = "pip install -r requirements.txt && python -m pytest"
		lintScript = "pip install flake8 black && flake8 . && black --check ."
	case "rust":
		buildScript = "cargo build --verbose"
		testScript = "cargo test --verbose"
		lintScript = "cargo clippy -- -D warnings && cargo fmt --check"
	case "java":
		buildScript = "./gradlew build"
		testScript = "./gradlew test"
		lintScript = "./gradlew check"
	default:
		buildScript = "echo 'Add build steps'"
		testScript = "echo 'Add test steps'"
		lintScript = "echo 'Add lint steps'"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`stages:
  - lint
  - build
  - test`))
	if addDocker {
		sb.WriteString("\n  - docker")
	}
	if addDeploy {
		sb.WriteString("\n  - deploy")
	}

	sb.WriteString(fmt.Sprintf(`

lint:
  stage: lint
  script:
    - %s

build:
  stage: build
  script:
    - %s

test:
  stage: test
  script:
    - %s
  coverage: '/coverage: \d+.\d+%%/'
`, lintScript, buildScript, testScript))

	if addDocker {
		sb.WriteString(`
docker:
  stage: docker
  image: docker:latest
  services:
    - docker:dind
  only:
    - main
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA -t $CI_REGISTRY_IMAGE:latest .
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
`)
	}

	if addDeploy {
		sb.WriteString(`
deploy:
  stage: deploy
  only:
    - main
  environment:
    name: production
  script:
    - echo "Add your deployment steps here"
    - echo "Examples: kubectl apply, helm upgrade, SSH deploy, etc."
`)
	}

	outPath := ".gitlab-ci.yml"
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		fmt.Printf("Error writing CI file: %v\n", err)
		return
	}
	fmt.Printf("✓ Generated GitLab CI config at %s\n", outPath)
}

func generateJenkinsfile(lang string, addDocker, addDeploy bool) {
	var buildCmd, testCmd, lintCmd string

	switch lang {
	case "go":
		buildCmd = "go build -v ./..."
		testCmd = "go test -v -race ./..."
		lintCmd = "golangci-lint run"
	case "node":
		buildCmd = "npm ci && npm run build --if-present"
		testCmd = "npm ci && npm test"
		lintCmd = "npm ci && npx eslint ."
	case "python":
		buildCmd = "pip install -r requirements.txt"
		testCmd = "python -m pytest"
		lintCmd = "flake8 . && black --check ."
	case "rust":
		buildCmd = "cargo build --verbose"
		testCmd = "cargo test --verbose"
		lintCmd = "cargo clippy -- -D warnings"
	case "java":
		buildCmd = "./gradlew build"
		testCmd = "./gradlew test"
		lintCmd = "./gradlew check"
	default:
		buildCmd = "echo 'Add build steps'"
		testCmd = "echo 'Add test steps'"
		lintCmd = "echo 'Add lint steps'"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`pipeline {
    agent any

    stages {
        stage('Lint') {
            steps {
                sh '%s'
            }
        }

        stage('Build') {
            steps {
                sh '%s'
            }
        }

        stage('Test') {
            steps {
                sh '%s'
            }
        }
`, lintCmd, buildCmd, testCmd))

	if addDocker {
		sb.WriteString(`
        stage('Docker Build & Push') {
            when {
                branch 'main'
            }
            steps {
                script {
                    def image = docker.build("${env.REGISTRY}/${env.IMAGE_NAME}:${env.BUILD_NUMBER}")
                    docker.withRegistry("https://${env.REGISTRY}", 'registry-credentials') {
                        image.push()
                        image.push('latest')
                    }
                }
            }
        }
`)
	}

	if addDeploy {
		sb.WriteString(`
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                echo 'Add your deployment steps here'
                echo 'Examples: kubectl apply, helm upgrade, SSH deploy, etc.'
            }
        }
`)
	}

	sb.WriteString(`    }

    post {
        always {
            cleanWs()
        }
        success {
            echo 'Pipeline succeeded!'
        }
        failure {
            echo 'Pipeline failed!'
        }
    }
}
`)

	outPath := "Jenkinsfile"
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		fmt.Printf("Error writing Jenkinsfile: %v\n", err)
		return
	}
	fmt.Printf("✓ Generated Jenkinsfile at %s\n", outPath)
}

func init() {
	projectCmd.AddCommand(ciCmd)
}
