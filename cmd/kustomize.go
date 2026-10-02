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

var kustomizeCmd = &cobra.Command{
	Use:   "kustomize",
	Short: "Scaffold Kustomize overlays for environment-based deployments",
	Long: `kustomize generates a directory structure with a base configuration and
environment overlays (dev, staging, production). This follows Kustomize best practices:

  k8s/
  ├── base/
  │   ├── kustomization.yaml
  │   ├── deployment.yaml
  │   └── service.yaml
  └── overlays/
      ├── dev/
      │   └── kustomization.yaml
      ├── staging/
      │   └── kustomization.yaml
      └── production/
          └── kustomization.yaml

Each overlay patches the base with environment-specific values (replicas, resources, etc.).`,
	Run: runKustomize,
}

func runKustomize(cmd *cobra.Command, args []string) {
	var appName, image, portStr, replicasStr string

	survey.AskOne(&survey.Input{Message: "Application name:", Default: "my-app"}, &appName)
	survey.AskOne(&survey.Input{Message: "Container image:", Default: "nginx:latest"}, &image)
	survey.AskOne(&survey.Input{Message: "Container port:", Default: "8080"}, &portStr)
	survey.AskOne(&survey.Input{Message: "Base replica count:", Default: "1"}, &replicasStr)

	appName = strings.TrimSpace(appName)
	image = strings.TrimSpace(image)
	port := strings.TrimSpace(portStr)
	replicas := strings.TrimSpace(replicasStr)

	baseDir := filepath.Join("k8s", "base")
	overlaysDir := filepath.Join("k8s", "overlays")

	// Create all directories
	envs := []string{"dev", "staging", "production"}
	dirs := []string{baseDir}
	for _, env := range envs {
		dirs = append(dirs, filepath.Join(overlaysDir, env))
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	generated := 0

	// === Base ===

	// base/deployment.yaml
	baseDeployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: %s
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
        - name: %s
          image: %s
          ports:
            - containerPort: %s
          resources:
            requests:
              cpu: 100m
              memory: 64Mi
            limits:
              cpu: 250m
              memory: 128Mi
          livenessProbe:
            httpGet:
              path: /
              port: %s
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /
              port: %s
            initialDelaySeconds: 5
            periodSeconds: 5
`, appName, appName, replicas, appName, appName, appName, image, port, port, port)
	writeKustomizeFile(baseDir, "deployment.yaml", baseDeployment)
	generated++

	// base/service.yaml
	baseService := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  labels:
    app: %s
spec:
  type: ClusterIP
  selector:
    app: %s
  ports:
    - protocol: TCP
      port: %s
      targetPort: %s
`, appName, appName, appName, port, port)
	writeKustomizeFile(baseDir, "service.yaml", baseService)
	generated++

	// base/kustomization.yaml
	baseKustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  app: %s

resources:
  - deployment.yaml
  - service.yaml
`, appName)
	writeKustomizeFile(baseDir, "kustomization.yaml", baseKustomization)
	generated++

	// === Overlays ===

	// dev overlay
	devKustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: %s-dev

resources:
  - ../../base

patches:
  - target:
      kind: Deployment
      name: %s
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 1
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/cpu
        value: 50m
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/memory
        value: 32Mi
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/cpu
        value: 100m
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/memory
        value: 64Mi

commonLabels:
  environment: dev
`, appName, appName)
	writeKustomizeFile(filepath.Join(overlaysDir, "dev"), "kustomization.yaml", devKustomization)
	generated++

	// staging overlay
	stagingKustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: %s-staging

resources:
  - ../../base

patches:
  - target:
      kind: Deployment
      name: %s
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 2
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/cpu
        value: 200m
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/memory
        value: 128Mi
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/cpu
        value: 500m
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/memory
        value: 256Mi

commonLabels:
  environment: staging
`, appName, appName)
	writeKustomizeFile(filepath.Join(overlaysDir, "staging"), "kustomization.yaml", stagingKustomization)
	generated++

	// production overlay
	prodKustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: %s-production

resources:
  - ../../base

patches:
  - target:
      kind: Deployment
      name: %s
    patch: |-
      - op: replace
        path: /spec/replicas
        value: 3
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/cpu
        value: 500m
      - op: replace
        path: /spec/template/spec/containers/0/resources/requests/memory
        value: 256Mi
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/cpu
        value: "1"
      - op: replace
        path: /spec/template/spec/containers/0/resources/limits/memory
        value: 512Mi

commonLabels:
  environment: production
`, appName, appName)
	writeKustomizeFile(filepath.Join(overlaysDir, "production"), "kustomization.yaml", prodKustomization)
	generated++

	fmt.Printf("\n✓ Generated Kustomize structure with %d files\n", generated)
	fmt.Printf("  Preview dev:    kubectl kustomize %s\n", filepath.ToSlash(filepath.Join(overlaysDir, "dev")))
	fmt.Printf("  Apply staging:  kubectl apply -k %s\n", filepath.ToSlash(filepath.Join(overlaysDir, "staging")))
	fmt.Printf("  Apply prod:     kubectl apply -k %s\n", filepath.ToSlash(filepath.Join(overlaysDir, "production")))
}

func writeKustomizeFile(dir, filename, content string) {
	outPath := filepath.Join(dir, filename)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("  ✓ %s\n", filepath.ToSlash(outPath))
}

func init() {
	clusterCmd.AddCommand(kustomizeCmd)
}
