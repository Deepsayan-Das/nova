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

var makefileCmd = &cobra.Command{
	Use:   "makefile",
	Short: "Generate a Makefile or Taskfile for your project",
	Long: `makefile generates a Makefile with common targets tailored to your project
language. Includes build, test, lint, clean, docker, and run targets.`,
	Run: runMakefile,
}

func runMakefile(cmd *cobra.Command, args []string) {
	var lang, appName string

	survey.AskOne(&survey.Select{
		Message: "Project language:",
		Options: []string{"go", "node", "python", "rust", "java", "generic"},
		Default: "go",
	}, &lang)

	survey.AskOne(&survey.Input{Message: "Application name:", Default: "my-app"}, &appName)
	appName = strings.TrimSpace(appName)

	content := generateMakefile(lang, appName)

	if err := os.WriteFile("Makefile", []byte(content), 0644); err != nil {
		fmt.Printf("Error writing Makefile: %v\n", err)
		return
	}
	fmt.Println("✓ Generated Makefile")
}

func generateMakefile(lang, appName string) string {
	header := fmt.Sprintf(`.PHONY: help build test lint clean run docker-build docker-run

APP_NAME := %s
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%%-20s\033[0m %%s\n", $$1, $$2}'

`, appName)

	switch lang {
	case "go":
		return header + fmt.Sprintf(`build: ## Build the Go binary
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o bin/$(APP_NAME) .

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out

run: build ## Build and run
	./bin/$(APP_NAME)

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 $(APP_NAME):latest

tidy: ## Tidy go modules
	go mod tidy
	go mod verify

all: lint vet test build ## Run all checks and build
`)

	case "node":
		return header + `build: ## Build the project
	npm run build

test: ## Run tests
	npm test

lint: ## Run linter
	npx eslint . --ext .js,.jsx,.ts,.tsx

fmt: ## Format code
	npx prettier --write .

clean: ## Clean build artifacts
	rm -rf node_modules dist build .next .nuxt

install: ## Install dependencies
	npm ci

run: ## Start development server
	npm run dev

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 3000:3000 $(APP_NAME):latest

all: install lint test build ## Run all checks and build
`

	case "python":
		return header + `VENV := .venv
PYTHON := $(VENV)/bin/python
PIP := $(VENV)/bin/pip

venv: ## Create virtual environment
	python3 -m venv $(VENV)
	$(PIP) install --upgrade pip

install: venv ## Install dependencies
	$(PIP) install -r requirements.txt
	$(PIP) install -r requirements-dev.txt 2>/dev/null || true

build: install ## Build/install the package
	$(PIP) install -e .

test: ## Run tests
	$(PYTHON) -m pytest -v --cov=. --cov-report=term-missing

lint: ## Run linters
	$(PYTHON) -m flake8 .
	$(PYTHON) -m mypy . || true

fmt: ## Format code
	$(PYTHON) -m black .
	$(PYTHON) -m isort .

clean: ## Clean build artifacts
	rm -rf $(VENV) __pycache__ .pytest_cache .mypy_cache htmlcov .coverage *.egg-info dist build

run: ## Run the application
	$(PYTHON) main.py

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 8000:8000 $(APP_NAME):latest

all: install lint test ## Run all checks
`

	case "rust":
		return header + `build: ## Build release binary
	cargo build --release

build-debug: ## Build debug binary
	cargo build

test: ## Run tests
	cargo test --verbose

lint: ## Run clippy linter
	cargo clippy -- -D warnings

fmt: ## Format code
	cargo fmt

fmt-check: ## Check formatting
	cargo fmt --check

clean: ## Clean build artifacts
	cargo clean

run: ## Build and run
	cargo run

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 $(APP_NAME):latest

all: fmt-check lint test build ## Run all checks and build
`

	case "java":
		return header + `build: ## Build the project
	./gradlew build

test: ## Run tests
	./gradlew test

lint: ## Run code checks
	./gradlew check

clean: ## Clean build artifacts
	./gradlew clean

run: ## Run the application
	./gradlew bootRun 2>/dev/null || ./gradlew run

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 $(APP_NAME):latest

all: lint test build ## Run all checks and build
`

	default:
		return header + `build: ## Build the project
	@echo "Add your build command here"

test: ## Run tests
	@echo "Add your test command here"

lint: ## Run linter
	@echo "Add your lint command here"

clean: ## Clean build artifacts
	@echo "Add your clean command here"

run: ## Run the application
	@echo "Add your run command here"

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run --rm $(APP_NAME):latest

all: lint test build ## Run all checks and build
`
	}
}

func init() {
	projectCmd.AddCommand(makefileCmd)
}
