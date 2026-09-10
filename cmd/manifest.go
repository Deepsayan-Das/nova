/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type K8sConfig struct {
	AppName       string `survey:"app_name"`
	Image         string `survey:"image"`
	Replicas      int    `survey:"replicas"`
	ContainerPort int    `survey:"container_port"`
	ServiceType   string `survey:"service_type"`
}

// GenerateK8sDeployment produces valid Kubernetes Deployment YAML
func GenerateK8sDeployment(cfg K8sConfig) string {
	if cfg.AppName == "" {
		cfg.AppName = "my-app"
	}
	if cfg.Replicas <= 0 {
		cfg.Replicas = 1
	}
	if cfg.ContainerPort <= 0 {
		cfg.ContainerPort = 80
	}

	depMap := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": cfg.AppName + "-deployment",
			"labels": map[string]string{
				"app": cfg.AppName,
			},
		},
		"spec": map[string]interface{}{
			"replicas": cfg.Replicas,
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{
					"app": cfg.AppName,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{
						"app": cfg.AppName,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  cfg.AppName,
							"image": cfg.Image,
							"ports": []map[string]interface{}{
								{
									"containerPort": cfg.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(depMap)
	if err != nil {
		return ""
	}
	return string(data)
}

// GenerateK8sService produces valid Kubernetes Service YAML
func GenerateK8sService(cfg K8sConfig) string {
	if cfg.AppName == "" {
		cfg.AppName = "my-app"
	}
	if cfg.ContainerPort <= 0 {
		cfg.ContainerPort = 80
	}
	if cfg.ServiceType == "" {
		cfg.ServiceType = "ClusterIP"
	}

	svcMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": cfg.AppName + "-service",
			"labels": map[string]string{
				"app": cfg.AppName,
			},
		},
		"spec": map[string]interface{}{
			"type": cfg.ServiceType,
			"selector": map[string]string{
				"app": cfg.AppName,
			},
			"ports": []map[string]interface{}{
				{
					"protocol":   "TCP",
					"port":       cfg.ContainerPort,
					"targetPort": cfg.ContainerPort,
				},
			},
		},
	}

	data, err := yaml.Marshal(svcMap)
	if err != nil {
		return ""
	}
	return string(data)
}

// GenerateK8sManifest returns combined deployment and service YAML strings
func GenerateK8sManifest(cfg K8sConfig) (string, string) {
	return GenerateK8sDeployment(cfg), GenerateK8sService(cfg)
}

// manifestCmd represents the manifest command under cluster
var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Generate Kubernetes Deployment and Service manifests",
	Long:  `Interactively prompt for Kubernetes app name, container image, replicas, ports, and service type, and write deployment.yaml and service.yaml inside ./k8s.`,
	Run: func(cmd *cobra.Command, args []string) {
		var qs = []*survey.Question{
			{
				Name: "app_name",
				Prompt: &survey.Input{
					Message: "App name:",
					Default: "my-app",
				},
			},
			{
				Name: "image",
				Prompt: &survey.Input{
					Message: "Container image (e.g. nginx:latest, myrepo/my-app:1.0):",
				},
			},
			{
				Name: "replicas",
				Prompt: &survey.Input{
					Message: "Replica count:",
					Default: "2",
				},
			},
			{
				Name: "container_port",
				Prompt: &survey.Input{
					Message: "Container port:",
					Default: "8080",
				},
			},
			{
				Name: "service_type",
				Prompt: &survey.Select{
					Message: "Kubernetes Service Type:",
					Options: []string{"ClusterIP", "NodePort", "LoadBalancer"},
					Default: "ClusterIP",
				},
			},
		}

		answers := struct {
			AppName       string `survey:"app_name"`
			Image         string `survey:"image"`
			Replicas      string `survey:"replicas"`
			ContainerPort string `survey:"container_port"`
			ServiceType   string `survey:"service_type"`
		}{}

		err := survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err)
			return
		}

		replicas, _ := strconv.Atoi(answers.Replicas)
		port, _ := strconv.Atoi(answers.ContainerPort)

		cfg := K8sConfig{
			AppName:       strings.TrimSpace(answers.AppName),
			Image:         strings.TrimSpace(answers.Image),
			Replicas:      replicas,
			ContainerPort: port,
			ServiceType:   answers.ServiceType,
		}

		depYaml, svcYaml := GenerateK8sManifest(cfg)

		k8sDir := "k8s"
		err = os.MkdirAll(k8sDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", k8sDir, err)
			return
		}

		depPath := filepath.Join(k8sDir, "deployment.yaml")
		svcPath := filepath.Join(k8sDir, "service.yaml")

		err = os.WriteFile(depPath, []byte(depYaml), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", depPath, err)
			return
		}

		err = os.WriteFile(svcPath, []byte(svcYaml), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", svcPath, err)
			return
		}

		fmt.Printf("✓ Kubernetes manifests generated at %s and %s\n", filepath.ToSlash(depPath), filepath.ToSlash(svcPath))
	},
}

func init() {
	clusterCmd.AddCommand(manifestCmd)
}
