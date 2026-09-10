package cmd

import (
	"strings"
	"testing"
)

func TestGenerateDockerfile_SingleStage(t *testing.T) {
	cfg := DockerfileConfig{
		BaseImage:  "node",
		Version:    "20-alpine",
		Port:       "3000",
		WorkDir:    "/app",
		EntryCmd:   "npm start",
		Multistage: false,
	}

	got := GenerateDockerfile(cfg)

	if !strings.Contains(got, "FROM node:20-alpine") {
		t.Errorf("Expected FROM node:20-alpine, got:\n%s", got)
	}
	if !strings.Contains(got, "WORKDIR /app") {
		t.Errorf("Expected WORKDIR /app, got:\n%s", got)
	}
	if !strings.Contains(got, "EXPOSE 3000") {
		t.Errorf("Expected EXPOSE 3000, got:\n%s", got)
	}
	if !strings.Contains(got, `CMD ["npm", "start"]`) {
		t.Errorf("Expected CMD [\"npm\", \"start\"], got:\n%s", got)
	}
}

func TestGenerateDockerfile_MultiStage_Go(t *testing.T) {
	cfg := DockerfileConfig{
		BaseImage:  "go",
		Version:    "1.22",
		Port:       "8080",
		WorkDir:    "/app",
		EntryCmd:   "./app",
		Multistage: true,
	}

	got := GenerateDockerfile(cfg)

	if !strings.Contains(got, "FROM golang:1.22 AS builder") {
		t.Errorf("Expected builder stage for Go, got:\n%s", got)
	}
	if !strings.Contains(got, "FROM alpine:latest") {
		t.Errorf("Expected runtime stage alpine:latest, got:\n%s", got)
	}
	if !strings.Contains(got, "COPY --from=builder /app/app ./app") {
		t.Errorf("Expected artifact copy from builder, got:\n%s", got)
	}
}

func TestGenerateDockerfile_MultiStage_Node(t *testing.T) {
	cfg := DockerfileConfig{
		BaseImage:  "node",
		Version:    "20",
		Port:       "3000",
		WorkDir:    "/app",
		EntryCmd:   "npm start",
		Multistage: true,
	}

	got := GenerateDockerfile(cfg)

	if !strings.Contains(got, "FROM node:20 AS builder") {
		t.Errorf("Expected node builder stage, got:\n%s", got)
	}
	if !strings.Contains(got, "FROM node:20-alpine") {
		t.Errorf("Expected node runtime stage, got:\n%s", got)
	}
	if !strings.Contains(got, "COPY --from=builder /app/dist ./dist") {
		t.Errorf("Expected dist copy from builder, got:\n%s", got)
	}
}

func TestGenerateDockerfile_MultiStage_PythonFallback(t *testing.T) {
	cfg := DockerfileConfig{
		BaseImage:  "python",
		Version:    "3.12-slim",
		Port:       "8000",
		WorkDir:    "/app",
		EntryCmd:   "python app.py",
		Multistage: true,
	}

	got := GenerateDockerfile(cfg)

	if !strings.Contains(got, "Note: Python projects typically use single-stage builds") {
		t.Errorf("Expected python single-stage note, got:\n%s", got)
	}
	if !strings.Contains(got, "FROM python:3.12-slim") {
		t.Errorf("Expected FROM python:3.12-slim, got:\n%s", got)
	}
}
