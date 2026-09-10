package cmd

import (
	"strings"
	"testing"
)

func TestGenerateCompose_SingleService(t *testing.T) {
	cfg := ComposeConfig{
		Version: "3.8",
		Services: map[string]ServiceConfig{
			"web": {
				Image: "node:20-alpine",
				Ports: []string{"3000:3000"},
				Environment: map[string]string{
					"NODE_ENV": "production",
				},
			},
		},
	}

	got := GenerateCompose(cfg)

	if !strings.Contains(got, `version: "3.8"`) && !strings.Contains(got, `version: '3.8'`) && !strings.Contains(got, `version: 3.8`) {
		t.Errorf("Expected version 3.8, got:\n%s", got)
	}
	if !strings.Contains(got, "web:") {
		t.Errorf("Expected web service block, got:\n%s", got)
	}
	if !strings.Contains(got, "node:20-alpine") {
		t.Errorf("Expected image node:20-alpine, got:\n%s", got)
	}
	if !strings.Contains(got, "3000:3000") {
		t.Errorf("Expected port 3000:3000, got:\n%s", got)
	}
	if !strings.Contains(got, "NODE_ENV: production") {
		t.Errorf("Expected env variable NODE_ENV, got:\n%s", got)
	}
}

func TestGenerateCompose_MultiService_WithDependsOn(t *testing.T) {
	cfg := ComposeConfig{
		Version: "3.8",
		Services: map[string]ServiceConfig{
			"api": {
				Build:     ".",
				Ports:     []string{"8080:8080"},
				DependsOn: []string{"db"},
			},
			"db": {
				Image: "postgres:15-alpine",
				Environment: map[string]string{
					"POSTGRES_PASSWORD": "secretpassword",
				},
			},
		},
	}

	got := GenerateCompose(cfg)

	if !strings.Contains(got, "api:") || !strings.Contains(got, "db:") {
		t.Errorf("Expected api and db services, got:\n%s", got)
	}
	if !strings.Contains(got, "depends_on:") {
		t.Errorf("Expected depends_on block, got:\n%s", got)
	}
	if !strings.Contains(got, "db") {
		t.Errorf("Expected dependency on db, got:\n%s", got)
	}
}
