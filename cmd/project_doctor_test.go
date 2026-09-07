package cmd

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNovaMetadata_UnmarshalWithVersion(t *testing.T) {
	yamlData := []byte(`
language: node
template: express
created: 2026-09-07
nova_version: 1.0.0
version: "20.11.0"
`)

	var meta NovaMetadata
	err := yaml.Unmarshal(yamlData, &meta)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with version: %v", err)
	}

	if meta.Language != "node" {
		t.Errorf("Expected language 'node', got '%s'", meta.Language)
	}
	if meta.Version != "20.11.0" {
		t.Errorf("Expected version '20.11.0', got '%s'", meta.Version)
	}
}

func TestNovaMetadata_UnmarshalWithoutVersion(t *testing.T) {
	yamlData := []byte(`
language: go
template: cli
created: 2026-09-07
nova_version: 1.0.0
`)

	var meta NovaMetadata
	err := yaml.Unmarshal(yamlData, &meta)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML without version: %v", err)
	}

	if meta.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", meta.Language)
	}
	if meta.Version != "" {
		t.Errorf("Expected empty version, got '%s'", meta.Version)
	}
}

func TestParseVersionString(t *testing.T) {
	tests := []struct {
		raw      string
		lang     string
		expected string
	}{
		{raw: "v20.11.0", lang: "node", expected: "20.11.0"},
		{raw: "20.11.0", lang: "node", expected: "20.11.0"},
		{raw: "Python 3.10.12", lang: "python", expected: "3.10.12"},
		{raw: "3.10.12", lang: "python", expected: "3.10.12"},
		{raw: "go version go1.22.1 linux/amd64", lang: "go", expected: "1.22.1"},
		{raw: "go1.22.1", lang: "go", expected: "1.22.1"},
		{raw: "rustc 1.75.0 (82e1608df 2023-12-21)", lang: "rust", expected: "1.75.0"},
	}

	for _, tt := range tests {
		got := parseVersionString(tt.raw, tt.lang)
		if got != tt.expected {
			t.Errorf("parseVersionString(%q, %q) = %q; expected %q", tt.raw, tt.lang, got, tt.expected)
		}
	}
}

func TestIsVersionMatched(t *testing.T) {
	tests := []struct {
		active   string
		pinned   string
		expected bool
	}{
		{active: "20.11.0", pinned: "20.11.0", expected: true},
		{active: "v20.11.0", pinned: "20.11.0", expected: true},
		{active: "20.11.0", pinned: "v20.11.0", expected: true},
		{active: "20.11.4", pinned: "20.11", expected: true},
		{active: "18.17.0", pinned: "20.11.0", expected: false},
		{active: "1.22.1", pinned: "1.24.0", expected: false},
	}

	for _, tt := range tests {
		got := isVersionMatched(tt.active, tt.pinned)
		if got != tt.expected {
			t.Errorf("isVersionMatched(%q, %q) = %v; expected %v", tt.active, tt.pinned, got, tt.expected)
		}
	}
}

func TestGetVersionFixAdvice(t *testing.T) {
	tests := []struct {
		lang     string
		pinned   string
		contains string
	}{
		{lang: "node", pinned: "20.11.0", contains: "nvm install 20.11.0 && nvm use 20.11.0"},
		{lang: "python", pinned: "3.10.12", contains: "pyenv install 3.10.12 && pyenv local 3.10.12"},
		{lang: "go", pinned: "1.22.1", contains: "g use 1.22.1"},
		{lang: "rust", pinned: "1.75.0", contains: "rustup override set 1.75.0"},
	}

	for _, tt := range tests {
		got := getVersionFixAdvice(tt.lang, tt.pinned)
		if got == "" {
			t.Errorf("getVersionFixAdvice(%q, %q) returned empty string", tt.lang, tt.pinned)
		}
	}
}
