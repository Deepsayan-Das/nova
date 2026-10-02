/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ContainerBuildConfig struct {
	Dockerfile  string
	Tag         string
	ContextPath string
}

type ContainerRunConfig struct {
	Image     string
	Ports     []string
	Env       []string
	Detach    bool
	Name      string
	ExtraArgs []string
}

type ContainerPsConfig struct {
	All bool
}

type ContainerLogsConfig struct {
	ContainerID string
	Follow      bool
}

// CheckPodmanInstalled verifies if podman is available on PATH.
func CheckPodmanInstalled() error {
	_, err := exec.LookPath("podman")
	if err != nil {
		return fmt.Errorf("'podman' executable not found on PATH.\nPlease install podman (e.g. run 'nova install podman' or use your system package manager)")
	}
	return nil
}

// BuildPodmanBuildArgs constructs argument slice for podman build
func BuildPodmanBuildArgs(cfg ContainerBuildConfig) []string {
	args := []string{"build"}
	if cfg.Dockerfile != "" {
		args = append(args, "-f", cfg.Dockerfile)
	}
	if cfg.Tag != "" {
		args = append(args, "-t", cfg.Tag)
	}
	ctx := cfg.ContextPath
	if ctx == "" {
		ctx = "."
	}
	args = append(args, ctx)
	return args
}

// BuildPodmanRunArgs constructs argument slice for podman run
func BuildPodmanRunArgs(cfg ContainerRunConfig) []string {
	args := []string{"run"}
	if cfg.Detach {
		args = append(args, "-d")
	}
	if cfg.Name != "" {
		args = append(args, "--name", cfg.Name)
	}
	for _, p := range cfg.Ports {
		p = strings.TrimSpace(p)
		if p != "" {
			args = append(args, "-p", p)
		}
	}
	for _, e := range cfg.Env {
		e = strings.TrimSpace(e)
		if e != "" {
			args = append(args, "-e", e)
		}
	}
	args = append(args, cfg.ExtraArgs...)
	if cfg.Image != "" {
		args = append(args, cfg.Image)
	}
	return args
}

// BuildPodmanPsArgs constructs argument slice for podman ps
func BuildPodmanPsArgs(cfg ContainerPsConfig) []string {
	args := []string{"ps"}
	if cfg.All {
		args = append(args, "-a")
	}
	return args
}

// BuildPodmanStopArgs constructs argument slice for podman stop
func BuildPodmanStopArgs(nameOrID string) []string {
	args := []string{"stop"}
	if nameOrID != "" {
		args = append(args, nameOrID)
	}
	return args
}

// BuildPodmanLogsArgs constructs argument slice for podman logs
func BuildPodmanLogsArgs(cfg ContainerLogsConfig) []string {
	args := []string{"logs"}
	if cfg.Follow {
		args = append(args, "-f")
	}
	if cfg.ContainerID != "" {
		args = append(args, cfg.ContainerID)
	}
	return args
}

// RunPodmanCommand executes podman with specified args.
// If streamOutput is true, stdout/stderr stream directly to terminal.
func RunPodmanCommand(args []string, streamOutput bool) (string, error) {
	if err := CheckPodmanInstalled(); err != nil {
		return "", err
	}

	cmd := exec.Command("podman", args...)

	if streamOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		return "", err
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	output := strings.TrimSpace(outBuf.String())
	if err != nil {
		stderrStr := strings.TrimSpace(errBuf.String())
		if stderrStr != "" {
			return output, fmt.Errorf("%v: %s", err, stderrStr)
		}
		return output, err
	}

	return output, nil
}
