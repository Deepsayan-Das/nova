package cmd

import (
	"testing"
)

func TestResolveAptPackage_RegistryShortcuts(t *testing.T) {
	tools := []InstallTool{
		{
			Name:        "Go",
			Key:         "go",
			Description: "Go programming language compiler",
			Executables: []string{"go"},
			AptPackage:  "golang",
		},
		{
			Name:        "Java (JDK)",
			Key:         "java",
			Description: "Java Development Kit",
			Executables: []string{"java", "javac"},
			AptPackage:  "default-jdk",
		},
	}

	tests := []struct {
		input            string
		expectedName     string
		expectedAptPkg   string
		expectedExecLen  int
	}{
		{input: "go", expectedName: "Go", expectedAptPkg: "golang", expectedExecLen: 1},
		{input: "GO", expectedName: "Go", expectedAptPkg: "golang", expectedExecLen: 1},
		{input: "java", expectedName: "Java (JDK)", expectedAptPkg: "default-jdk", expectedExecLen: 2},
	}

	for _, tt := range tests {
		name, aptPkg, exes := resolveAptPackage(tt.input, tools)
		if name != tt.expectedName {
			t.Errorf("For input '%s', expected name '%s', got '%s'", tt.input, tt.expectedName, name)
		}
		if aptPkg != tt.expectedAptPkg {
			t.Errorf("For input '%s', expected apt package '%s', got '%s'", tt.input, tt.expectedAptPkg, aptPkg)
		}
		if len(exes) != tt.expectedExecLen {
			t.Errorf("For input '%s', expected %d executables, got %d", tt.input, tt.expectedExecLen, len(exes))
		}
	}
}

func TestResolveAptPackage_ArbitraryAptPackage(t *testing.T) {
	tools := []InstallTool{
		{
			Name:        "Go",
			Key:         "go",
			AptPackage:  "golang",
			Executables: []string{"go"},
		},
	}

	input := "htop"
	name, aptPkg, exes := resolveAptPackage(input, tools)

	if name != "htop" {
		t.Errorf("Expected name 'htop', got '%s'", name)
	}
	if aptPkg != "htop" {
		t.Errorf("Expected aptPackage 'htop', got '%s'", aptPkg)
	}
	if len(exes) != 1 || exes[0] != "htop" {
		t.Errorf("Expected executables ['htop'], got %v", exes)
	}
}

func TestLoadInstallRegistry(t *testing.T) {
	tools := loadInstallRegistry()
	if len(tools) == 0 {
		t.Fatal("Expected non-empty install registry from default install.yaml")
	}

	foundGo := false
	for _, tool := range tools {
		if tool.Key == "go" {
			foundGo = true
			if tool.AptPackage != "golang" {
				t.Errorf("Expected apt package 'golang' for go tool key, got '%s'", tool.AptPackage)
			}
		}
	}

	if !foundGo {
		t.Error("Expected 'go' tool key in default install registry")
	}
}
