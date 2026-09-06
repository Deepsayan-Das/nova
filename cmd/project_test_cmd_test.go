package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolveTestRunner_Java_GradleWrapper(t *testing.T) {
	tempDir := t.TempDir()

	if runtime.GOOS == "windows" {
		batFile := filepath.Join(tempDir, "gradlew.bat")
		if err := os.WriteFile(batFile, []byte("@echo off"), 0755); err != nil {
			t.Fatalf("Failed to create gradlew.bat: %v", err)
		}

		execName, execArgs, err := resolveTestRunner("java", tempDir, []string{"--info"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if execName != "cmd.exe" {
			t.Errorf("Expected execName to be cmd.exe, got %s", execName)
		}
		expectedArgs := []string{"/c", "gradlew.bat", "test", "--info"}
		if len(execArgs) != len(expectedArgs) {
			t.Fatalf("Expected args count %d, got %d (%v)", len(expectedArgs), len(execArgs), execArgs)
		}
		for i, arg := range expectedArgs {
			if execArgs[i] != arg {
				t.Errorf("Expected arg[%d]=%s, got %s", i, arg, execArgs[i])
			}
		}
	} else {
		shFile := filepath.Join(tempDir, "gradlew")
		if err := os.WriteFile(shFile, []byte("#!/bin/sh"), 0755); err != nil {
			t.Fatalf("Failed to create gradlew: %v", err)
		}

		execName, execArgs, err := resolveTestRunner("java", tempDir, []string{"--info"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		absExpected, _ := filepath.Abs(shFile)
		if execName != absExpected {
			t.Errorf("Expected execName to be %s, got %s", absExpected, execName)
		}
		if len(execArgs) != 2 || execArgs[0] != "test" || execArgs[1] != "--info" {
			t.Errorf("Unexpected execArgs: %v", execArgs)
		}
	}
}

func TestResolveTestRunner_Java_PomXml(t *testing.T) {
	tempDir := t.TempDir()
	pomFile := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomFile, []byte("<project></project>"), 0644); err != nil {
		t.Fatalf("Failed to create pom.xml: %v", err)
	}

	execName, execArgs, err := resolveTestRunner("java", tempDir, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if execName != "mvn" {
		t.Errorf("Expected execName to be 'mvn', got '%s'", execName)
	}
	if len(execArgs) != 1 || execArgs[0] != "test" {
		t.Errorf("Expected execArgs to be ['test'], got %v", execArgs)
	}
}

func TestResolveTestRunner_Java_BuildGradle(t *testing.T) {
	tempDir := t.TempDir()
	gradleFile := filepath.Join(tempDir, "build.gradle")
	if err := os.WriteFile(gradleFile, []byte("// gradle"), 0644); err != nil {
		t.Fatalf("Failed to create build.gradle: %v", err)
	}

	execName, execArgs, err := resolveTestRunner("java", tempDir, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if execName != "gradle" {
		t.Errorf("Expected execName to be 'gradle', got '%s'", execName)
	}
	if len(execArgs) != 1 || execArgs[0] != "test" {
		t.Errorf("Expected execArgs to be ['test'], got %v", execArgs)
	}
}

func TestResolveTestRunner_Java_BuildGradleKts(t *testing.T) {
	tempDir := t.TempDir()
	gradleFile := filepath.Join(tempDir, "build.gradle.kts")
	if err := os.WriteFile(gradleFile, []byte("// gradle kts"), 0644); err != nil {
		t.Fatalf("Failed to create build.gradle.kts: %v", err)
	}

	execName, execArgs, err := resolveTestRunner("java", tempDir, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if execName != "gradle" {
		t.Errorf("Expected execName to be 'gradle', got '%s'", execName)
	}
	if len(execArgs) != 1 || execArgs[0] != "test" {
		t.Errorf("Expected execArgs to be ['test'], got %v", execArgs)
	}
}

func TestResolveTestRunner_Java_Fallback(t *testing.T) {
	tempDir := t.TempDir()

	execName, execArgs, err := resolveTestRunner("java", tempDir, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if execName != "mvn" {
		t.Errorf("Expected fallback execName to be 'mvn', got '%s'", execName)
	}
	if len(execArgs) != 1 || execArgs[0] != "test" {
		t.Errorf("Expected execArgs to be ['test'], got %v", execArgs)
	}
}

func TestResolveTestRunner_Go(t *testing.T) {
	execName, execArgs, err := resolveTestRunner("go", "./", []string{"-v"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if execName != "go" {
		t.Errorf("Expected 'go', got '%s'", execName)
	}
	if strings.Join(execArgs, " ") != "test ./... -v" {
		t.Errorf("Unexpected args: %v", execArgs)
	}
}
