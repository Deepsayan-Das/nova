package cmd

import (
	"reflect"
	"testing"
)

func TestBuildPodmanBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ContainerBuildConfig
		expected []string
	}{
		{
			name: "Default build args",
			cfg: ContainerBuildConfig{
				Dockerfile:  "Dockerfile",
				Tag:         "my-app:1.0",
				ContextPath: ".",
			},
			expected: []string{"build", "-f", "Dockerfile", "-t", "my-app:1.0", "."},
		},
		{
			name: "Custom dockerfile and context",
			cfg: ContainerBuildConfig{
				Dockerfile:  "docker/Dockerfile.prod",
				Tag:         "prod-api:latest",
				ContextPath: "./src",
			},
			expected: []string{"build", "-f", "docker/Dockerfile.prod", "-t", "prod-api:latest", "./src"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPodmanBuildArgs(tt.cfg)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("BuildPodmanBuildArgs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildPodmanRunArgs(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ContainerRunConfig
		expected []string
	}{
		{
			name: "Foreground run with port and env",
			cfg: ContainerRunConfig{
				Image:  "node:20-alpine",
				Ports:  []string{"3000:3000"},
				Env:    []string{"NODE_ENV=production"},
				Detach: false,
				Name:   "web-server",
			},
			expected: []string{"run", "--name", "web-server", "-p", "3000:3000", "-e", "NODE_ENV=production", "node:20-alpine"},
		},
		{
			name: "Detached run without name",
			cfg: ContainerRunConfig{
				Image:  "redis:alpine",
				Ports:  []string{"6379:6379"},
				Detach: true,
			},
			expected: []string{"run", "-d", "-p", "6379:6379", "redis:alpine"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPodmanRunArgs(tt.cfg)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("BuildPodmanRunArgs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildPodmanPsArgs(t *testing.T) {
	gotRunning := BuildPodmanPsArgs(ContainerPsConfig{All: false})
	wantRunning := []string{"ps"}
	if !reflect.DeepEqual(gotRunning, wantRunning) {
		t.Errorf("BuildPodmanPsArgs(All: false) = %v, want %v", gotRunning, wantRunning)
	}

	gotAll := BuildPodmanPsArgs(ContainerPsConfig{All: true})
	wantAll := []string{"ps", "-a"}
	if !reflect.DeepEqual(gotAll, wantAll) {
		t.Errorf("BuildPodmanPsArgs(All: true) = %v, want %v", gotAll, wantAll)
	}
}

func TestBuildPodmanStopArgs(t *testing.T) {
	got := BuildPodmanStopArgs("my-container-id")
	want := []string{"stop", "my-container-id"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildPodmanStopArgs() = %v, want %v", got, want)
	}
}

func TestBuildPodmanLogsArgs(t *testing.T) {
	gotFollow := BuildPodmanLogsArgs(ContainerLogsConfig{ContainerID: "web-app", Follow: true})
	wantFollow := []string{"logs", "-f", "web-app"}
	if !reflect.DeepEqual(gotFollow, wantFollow) {
		t.Errorf("BuildPodmanLogsArgs(Follow: true) = %v, want %v", gotFollow, wantFollow)
	}

	gotStatic := BuildPodmanLogsArgs(ContainerLogsConfig{ContainerID: "web-app", Follow: false})
	wantStatic := []string{"logs", "web-app"}
	if !reflect.DeepEqual(gotStatic, wantStatic) {
		t.Errorf("BuildPodmanLogsArgs(Follow: false) = %v, want %v", gotStatic, wantStatic)
	}
}
