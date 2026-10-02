/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
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
)

// scaffoldCmd generates a complete Kubernetes application stack in one shot
var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate a complete Kubernetes application stack",
	Long: `scaffold creates a full set of Kubernetes manifests for a production-ready
application in a single command. It generates:
  - Namespace
  - Deployment
  - Service (ClusterIP/NodePort/LoadBalancer)
  - Ingress (optional)
  - ConfigMap (optional)
  - HPA (optional)
  - PersistentVolumeClaim (optional)
  - NetworkPolicy (optional)

All files are written to ./k8s/<app-name>/ for clean organization.`,
	Run: runScaffold,
}

func runScaffold(cmd *cobra.Command, args []string) {
	// Core config
	var appName, namespace, image, replicasStr, portStr, svcType string
	survey.AskOne(&survey.Input{Message: "Application name:", Default: "my-app"}, &appName)
	survey.AskOne(&survey.Input{Message: "Namespace:", Default: appName}, &namespace)
	survey.AskOne(&survey.Input{Message: "Container image:", Default: "nginx:latest"}, &image)
	survey.AskOne(&survey.Input{Message: "Replica count:", Default: "2"}, &replicasStr)
	survey.AskOne(&survey.Input{Message: "Container port:", Default: "8080"}, &portStr)
	survey.AskOne(&survey.Select{
		Message: "Service type:",
		Options: []string{"ClusterIP", "NodePort", "LoadBalancer"},
		Default: "ClusterIP",
	}, &svcType)

	appName = strings.TrimSpace(appName)
	namespace = strings.TrimSpace(namespace)
	image = strings.TrimSpace(image)
	replicas, _ := strconv.Atoi(strings.TrimSpace(replicasStr))
	port, _ := strconv.Atoi(strings.TrimSpace(portStr))

	// Optional components
	var addIngress, addConfigMap, addHPA, addPVC, addNetPol bool
	survey.AskOne(&survey.Confirm{Message: "Generate Ingress?", Default: true}, &addIngress)
	survey.AskOne(&survey.Confirm{Message: "Generate ConfigMap?", Default: true}, &addConfigMap)
	survey.AskOne(&survey.Confirm{Message: "Generate HPA (autoscaling)?", Default: true}, &addHPA)
	survey.AskOne(&survey.Confirm{Message: "Generate PersistentVolumeClaim?", Default: false}, &addPVC)
	survey.AskOne(&survey.Confirm{Message: "Generate NetworkPolicy?", Default: false}, &addNetPol)

	// Output directory
	k8sDir := filepath.Join("k8s", appName)
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", k8sDir, err)
		return
	}

	generated := 0

	// 1. Namespace
	nsContent := GenerateNamespace(namespace, map[string]string{"app": appName})
	writeScaffoldFile(k8sDir, "namespace.yaml", nsContent)
	generated++

	// 2. Deployment
	depCfg := K8sConfig{
		AppName:       appName,
		Image:         image,
		Replicas:      replicas,
		ContainerPort: port,
	}
	depContent := GenerateK8sDeployment(depCfg)
	writeScaffoldFile(k8sDir, "deployment.yaml", depContent)
	generated++

	// 3. Service
	svcCfg := K8sConfig{
		AppName:       appName,
		Image:         image,
		ContainerPort: port,
		ServiceType:   svcType,
	}
	svcContent := GenerateK8sService(svcCfg)
	writeScaffoldFile(k8sDir, "service.yaml", svcContent)
	generated++

	// 4. Ingress
	if addIngress {
		var host string
		survey.AskOne(&survey.Input{Message: "Ingress hostname:", Default: appName + ".example.com"}, &host)

		ingCfg := IngressConfig{
			Name:         appName + "-ingress",
			Namespace:    namespace,
			Host:         strings.TrimSpace(host),
			Path:         "/",
			ServiceName:  appName + "-service",
			ServicePort:  port,
			IngressClass: "nginx",
		}
		writeScaffoldFile(k8sDir, "ingress.yaml", GenerateIngress(ingCfg))
		generated++
	}

	// 5. ConfigMap
	if addConfigMap {
		cmCfg := ConfigMapConfig{
			Name:      appName + "-config",
			Namespace: namespace,
			Data: map[string]string{
				"APP_NAME":    appName,
				"APP_ENV":     "production",
				"LOG_LEVEL":   "info",
			},
		}
		writeScaffoldFile(k8sDir, "configmap.yaml", GenerateConfigMap(cmCfg))
		generated++
	}

	// 6. HPA
	if addHPA {
		hpaCfg := HPAConfig{
			Name:             appName + "-hpa",
			Namespace:        namespace,
			TargetDeployment: appName + "-deployment",
			MinReplicas:      replicas,
			MaxReplicas:      replicas * 5,
			CPUPercent:       70,
		}
		writeScaffoldFile(k8sDir, "hpa.yaml", GenerateHPA(hpaCfg))
		generated++
	}

	// 7. PVC
	if addPVC {
		var storageSize string
		survey.AskOne(&survey.Input{Message: "PVC storage size:", Default: "5Gi"}, &storageSize)

		pvcCfg := PVCConfig{
			Name:       appName + "-pvc",
			Namespace:  namespace,
			AccessMode: "ReadWriteOnce",
			Size:       strings.TrimSpace(storageSize),
		}
		writeScaffoldFile(k8sDir, "pvc.yaml", GeneratePVC(pvcCfg))
		generated++
	}

	// 8. NetworkPolicy
	if addNetPol {
		npCfg := NetworkPolicyConfig{
			Name:        appName + "-netpol",
			Namespace:   namespace,
			PodLabel:    appName,
			IngressFrom: "same-namespace",
			EgressTo:    "all",
			Port:        port,
		}
		writeScaffoldFile(k8sDir, "networkpolicy.yaml", GenerateNetworkPolicy(npCfg))
		generated++
	}

	fmt.Printf("\n✓ Generated %d Kubernetes manifests in %s/\n", generated, filepath.ToSlash(k8sDir))
	fmt.Printf("  Apply with: kubectl apply -f %s/\n", filepath.ToSlash(k8sDir))
}

func writeScaffoldFile(dir, filename, content string) {
	outPath := filepath.Join(dir, filename)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("  ✓ %s\n", filepath.ToSlash(outPath))
}

func init() {
	clusterCmd.AddCommand(scaffoldCmd)
}
