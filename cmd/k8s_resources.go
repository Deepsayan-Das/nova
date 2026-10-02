/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// Ingress
// ============================================================================

type IngressConfig struct {
	Name        string
	Namespace   string
	Host        string
	Path        string
	ServiceName string
	ServicePort int
	TLSEnabled  bool
	TLSSecret   string
	IngressClass string
}

func GenerateIngress(cfg IngressConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-ingress"
	}
	if cfg.Path == "" {
		cfg.Path = "/"
	}
	if cfg.ServicePort <= 0 {
		cfg.ServicePort = 80
	}
	if cfg.IngressClass == "" {
		cfg.IngressClass = "nginx"
	}

	pathType := "Prefix"

	ingressMap := map[string]interface{}{
		"apiVersion": "networking.k8s.io/v1",
		"kind":       "Ingress",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
			"annotations": map[string]string{
				"kubernetes.io/ingress.class": cfg.IngressClass,
			},
		},
		"spec": map[string]interface{}{
			"ingressClassName": cfg.IngressClass,
			"rules": []map[string]interface{}{
				{
					"host": cfg.Host,
					"http": map[string]interface{}{
						"paths": []map[string]interface{}{
							{
								"path":     cfg.Path,
								"pathType": pathType,
								"backend": map[string]interface{}{
									"service": map[string]interface{}{
										"name": cfg.ServiceName,
										"port": map[string]interface{}{
											"number": cfg.ServicePort,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if cfg.Namespace != "" {
		ingressMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	if cfg.TLSEnabled && cfg.TLSSecret != "" {
		ingressMap["spec"].(map[string]interface{})["tls"] = []map[string]interface{}{
			{
				"hosts":      []string{cfg.Host},
				"secretName": cfg.TLSSecret,
			},
		}
	}

	data, err := yaml.Marshal(ingressMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var ingressCmd = &cobra.Command{
	Use:   "ingress",
	Short: "Generate a Kubernetes Ingress manifest",
	Long:  `Interactively generate a Kubernetes Ingress resource with host-based routing, TLS, and path configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		answers := struct {
			Name         string `survey:"name"`
			Namespace    string `survey:"namespace"`
			Host         string `survey:"host"`
			Path         string `survey:"path"`
			ServiceName  string `survey:"service_name"`
			ServicePort  string `survey:"service_port"`
			IngressClass string `survey:"ingress_class"`
			TLSEnabled   bool   `survey:"tls_enabled"`
			TLSSecret    string `survey:"tls_secret"`
		}{}

		qs := []*survey.Question{
			{Name: "name", Prompt: &survey.Input{Message: "Ingress name:", Default: "my-app-ingress"}},
			{Name: "namespace", Prompt: &survey.Input{Message: "Namespace (blank for default):"}},
			{Name: "host", Prompt: &survey.Input{Message: "Hostname (e.g. app.example.com):", Default: "app.example.com"}},
			{Name: "path", Prompt: &survey.Input{Message: "Path:", Default: "/"}},
			{Name: "service_name", Prompt: &survey.Input{Message: "Backend service name:", Default: "my-app-service"}},
			{Name: "service_port", Prompt: &survey.Input{Message: "Backend service port:", Default: "80"}},
			{Name: "ingress_class", Prompt: &survey.Select{Message: "Ingress class:", Options: []string{"nginx", "traefik", "haproxy", "alb", "gce"}, Default: "nginx"}},
			{Name: "tls_enabled", Prompt: &survey.Confirm{Message: "Enable TLS?", Default: false}},
		}

		if err := survey.Ask(qs, &answers); err != nil {
			fmt.Println(err)
			return
		}

		if answers.TLSEnabled {
			survey.AskOne(&survey.Input{Message: "TLS secret name:", Default: answers.Name + "-tls"}, &answers.TLSSecret)
		}

		port, _ := strconv.Atoi(answers.ServicePort)
		cfg := IngressConfig{
			Name:         answers.Name,
			Namespace:    answers.Namespace,
			Host:         answers.Host,
			Path:         answers.Path,
			ServiceName:  answers.ServiceName,
			ServicePort:  port,
			IngressClass: answers.IngressClass,
			TLSEnabled:   answers.TLSEnabled,
			TLSSecret:    answers.TLSSecret,
		}

		content := GenerateIngress(cfg)
		writeK8sFile("ingress.yaml", content)
	},
}

// ============================================================================
// ConfigMap
// ============================================================================

type ConfigMapConfig struct {
	Name      string
	Namespace string
	Data      map[string]string
}

func GenerateConfigMap(cfg ConfigMapConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-config"
	}

	cmMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"data": cfg.Data,
	}

	if cfg.Namespace != "" {
		cmMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(cmMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var configmapCmd = &cobra.Command{
	Use:   "configmap",
	Short: "Generate a Kubernetes ConfigMap manifest",
	Long:  `Interactively create a ConfigMap with key-value pairs for application configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, countStr string
		survey.AskOne(&survey.Input{Message: "ConfigMap name:", Default: "my-app-config"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "How many data entries?", Default: "2"}, &countStr)

		count, _ := strconv.Atoi(strings.TrimSpace(countStr))
		if count < 1 {
			count = 1
		}

		data := make(map[string]string)
		for i := 0; i < count; i++ {
			var key, value string
			survey.AskOne(&survey.Input{Message: fmt.Sprintf("Key #%d:", i+1)}, &key)
			survey.AskOne(&survey.Input{Message: fmt.Sprintf("Value for '%s':", key)}, &value)
			if key != "" {
				data[strings.TrimSpace(key)] = value
			}
		}

		cfg := ConfigMapConfig{Name: strings.TrimSpace(name), Namespace: strings.TrimSpace(namespace), Data: data}
		content := GenerateConfigMap(cfg)
		writeK8sFile("configmap.yaml", content)
	},
}

// ============================================================================
// Secret
// ============================================================================

type SecretConfig struct {
	Name      string
	Namespace string
	Type      string
	Data      map[string]string
}

func GenerateSecret(cfg SecretConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-secret"
	}
	if cfg.Type == "" {
		cfg.Type = "Opaque"
	}

	// Base64-encode values
	encoded := make(map[string]string)
	for k, v := range cfg.Data {
		encoded[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}

	secretMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"type": cfg.Type,
		"data": encoded,
	}

	if cfg.Namespace != "" {
		secretMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(secretMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Generate a Kubernetes Secret manifest",
	Long:  `Interactively create a Kubernetes Secret with base64-encoded key-value data.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, secretType, countStr string
		survey.AskOne(&survey.Input{Message: "Secret name:", Default: "my-app-secret"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Select{
			Message: "Secret type:",
			Options: []string{"Opaque", "kubernetes.io/tls", "kubernetes.io/dockerconfigjson", "kubernetes.io/basic-auth"},
			Default: "Opaque",
		}, &secretType)
		survey.AskOne(&survey.Input{Message: "How many data entries?", Default: "2"}, &countStr)

		count, _ := strconv.Atoi(strings.TrimSpace(countStr))
		if count < 1 {
			count = 1
		}

		data := make(map[string]string)
		for i := 0; i < count; i++ {
			var key, value string
			survey.AskOne(&survey.Input{Message: fmt.Sprintf("Key #%d:", i+1)}, &key)
			survey.AskOne(&survey.Password{Message: fmt.Sprintf("Value for '%s' (will be base64-encoded):", key)}, &value)
			if key != "" {
				data[strings.TrimSpace(key)] = value
			}
		}

		cfg := SecretConfig{Name: strings.TrimSpace(name), Namespace: strings.TrimSpace(namespace), Type: secretType, Data: data}
		content := GenerateSecret(cfg)
		writeK8sFile("secret.yaml", content)
	},
}

// ============================================================================
// HorizontalPodAutoscaler
// ============================================================================

type HPAConfig struct {
	Name            string
	Namespace       string
	TargetDeployment string
	MinReplicas     int
	MaxReplicas     int
	CPUPercent      int
	MemoryPercent   int
}

func GenerateHPA(cfg HPAConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-hpa"
	}
	if cfg.MinReplicas <= 0 {
		cfg.MinReplicas = 1
	}
	if cfg.MaxReplicas <= 0 {
		cfg.MaxReplicas = 10
	}
	if cfg.CPUPercent <= 0 {
		cfg.CPUPercent = 70
	}

	metrics := []map[string]interface{}{
		{
			"type": "Resource",
			"resource": map[string]interface{}{
				"name": "cpu",
				"target": map[string]interface{}{
					"type":               "Utilization",
					"averageUtilization": cfg.CPUPercent,
				},
			},
		},
	}

	if cfg.MemoryPercent > 0 {
		metrics = append(metrics, map[string]interface{}{
			"type": "Resource",
			"resource": map[string]interface{}{
				"name": "memory",
				"target": map[string]interface{}{
					"type":               "Utilization",
					"averageUtilization": cfg.MemoryPercent,
				},
			},
		})
	}

	hpaMap := map[string]interface{}{
		"apiVersion": "autoscaling/v2",
		"kind":       "HorizontalPodAutoscaler",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"spec": map[string]interface{}{
			"scaleTargetRef": map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"name":       cfg.TargetDeployment,
			},
			"minReplicas": cfg.MinReplicas,
			"maxReplicas": cfg.MaxReplicas,
			"metrics":     metrics,
		},
	}

	if cfg.Namespace != "" {
		hpaMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(hpaMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var hpaCmd = &cobra.Command{
	Use:   "hpa",
	Short: "Generate a Kubernetes HorizontalPodAutoscaler manifest",
	Long:  `Interactively create an HPA with CPU/memory scaling targets for a Deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, target, minStr, maxStr, cpuStr, memStr string
		var addMemory bool

		survey.AskOne(&survey.Input{Message: "HPA name:", Default: "my-app-hpa"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Target Deployment name:", Default: "my-app-deployment"}, &target)
		survey.AskOne(&survey.Input{Message: "Min replicas:", Default: "1"}, &minStr)
		survey.AskOne(&survey.Input{Message: "Max replicas:", Default: "10"}, &maxStr)
		survey.AskOne(&survey.Input{Message: "Target CPU utilization %:", Default: "70"}, &cpuStr)
		survey.AskOne(&survey.Confirm{Message: "Add memory-based scaling?", Default: false}, &addMemory)

		memPct := 0
		if addMemory {
			survey.AskOne(&survey.Input{Message: "Target Memory utilization %:", Default: "80"}, &memStr)
			memPct, _ = strconv.Atoi(strings.TrimSpace(memStr))
		}

		minR, _ := strconv.Atoi(strings.TrimSpace(minStr))
		maxR, _ := strconv.Atoi(strings.TrimSpace(maxStr))
		cpuPct, _ := strconv.Atoi(strings.TrimSpace(cpuStr))

		cfg := HPAConfig{
			Name:            strings.TrimSpace(name),
			Namespace:       strings.TrimSpace(namespace),
			TargetDeployment: strings.TrimSpace(target),
			MinReplicas:     minR,
			MaxReplicas:     maxR,
			CPUPercent:      cpuPct,
			MemoryPercent:   memPct,
		}
		content := GenerateHPA(cfg)
		writeK8sFile("hpa.yaml", content)
	},
}

// ============================================================================
// CronJob
// ============================================================================

type CronJobConfig struct {
	Name      string
	Namespace string
	Schedule  string
	Image     string
	Command   []string
	Suspend   bool
}

func GenerateCronJob(cfg CronJobConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-cronjob"
	}
	if cfg.Schedule == "" {
		cfg.Schedule = "*/5 * * * *"
	}

	container := map[string]interface{}{
		"name":  cfg.Name,
		"image": cfg.Image,
	}
	if len(cfg.Command) > 0 {
		container["command"] = cfg.Command
	}

	cronMap := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "CronJob",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"spec": map[string]interface{}{
			"schedule": cfg.Schedule,
			"suspend":  cfg.Suspend,
			"jobTemplate": map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers":    []map[string]interface{}{container},
							"restartPolicy": "OnFailure",
						},
					},
				},
			},
		},
	}

	if cfg.Namespace != "" {
		cronMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(cronMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var cronjobCmd = &cobra.Command{
	Use:   "cronjob",
	Short: "Generate a Kubernetes CronJob manifest",
	Long:  `Interactively create a CronJob with schedule, image, and command configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, schedule, image, cmdStr string
		var suspend bool

		survey.AskOne(&survey.Input{Message: "CronJob name:", Default: "my-cronjob"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Cron schedule (e.g. '*/5 * * * *'):", Default: "0 * * * *"}, &schedule)
		survey.AskOne(&survey.Input{Message: "Container image:", Default: "busybox:latest"}, &image)
		survey.AskOne(&survey.Input{Message: "Command (space-separated, e.g. '/bin/sh -c echo hello'):"}, &cmdStr)
		survey.AskOne(&survey.Confirm{Message: "Start suspended?", Default: false}, &suspend)

		var command []string
		if strings.TrimSpace(cmdStr) != "" {
			command = strings.Fields(cmdStr)
		}

		cfg := CronJobConfig{
			Name:      strings.TrimSpace(name),
			Namespace: strings.TrimSpace(namespace),
			Schedule:  strings.TrimSpace(schedule),
			Image:     strings.TrimSpace(image),
			Command:   command,
			Suspend:   suspend,
		}
		content := GenerateCronJob(cfg)
		writeK8sFile("cronjob.yaml", content)
	},
}

// ============================================================================
// PersistentVolumeClaim
// ============================================================================

type PVCConfig struct {
	Name         string
	Namespace    string
	StorageClass string
	AccessMode   string
	Size         string
}

func GeneratePVC(cfg PVCConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-pvc"
	}
	if cfg.AccessMode == "" {
		cfg.AccessMode = "ReadWriteOnce"
	}
	if cfg.Size == "" {
		cfg.Size = "1Gi"
	}

	pvcMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"spec": map[string]interface{}{
			"accessModes": []string{cfg.AccessMode},
			"resources": map[string]interface{}{
				"requests": map[string]string{
					"storage": cfg.Size,
				},
			},
		},
	}

	if cfg.Namespace != "" {
		pvcMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	if cfg.StorageClass != "" {
		pvcMap["spec"].(map[string]interface{})["storageClassName"] = cfg.StorageClass
	}

	data, err := yaml.Marshal(pvcMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var pvcCmd = &cobra.Command{
	Use:   "pvc",
	Short: "Generate a Kubernetes PersistentVolumeClaim manifest",
	Long:  `Interactively create a PVC with storage class, access mode, and size configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, storageClass, accessMode, size string

		survey.AskOne(&survey.Input{Message: "PVC name:", Default: "my-app-pvc"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Storage class (blank for cluster default):"}, &storageClass)
		survey.AskOne(&survey.Select{
			Message: "Access mode:",
			Options: []string{"ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany"},
			Default: "ReadWriteOnce",
		}, &accessMode)
		survey.AskOne(&survey.Input{Message: "Storage size (e.g. 1Gi, 10Gi, 100Gi):", Default: "1Gi"}, &size)

		cfg := PVCConfig{
			Name:         strings.TrimSpace(name),
			Namespace:    strings.TrimSpace(namespace),
			StorageClass: strings.TrimSpace(storageClass),
			AccessMode:   accessMode,
			Size:         strings.TrimSpace(size),
		}
		content := GeneratePVC(cfg)
		writeK8sFile("pvc.yaml", content)
	},
}

// ============================================================================
// StatefulSet
// ============================================================================

type StatefulSetConfig struct {
	Name          string
	Namespace     string
	Image         string
	Replicas      int
	ContainerPort int
	ServiceName   string
	StorageSize   string
}

func GenerateStatefulSet(cfg StatefulSetConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app"
	}
	if cfg.Replicas <= 0 {
		cfg.Replicas = 3
	}
	if cfg.ContainerPort <= 0 {
		cfg.ContainerPort = 80
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = cfg.Name + "-service"
	}
	if cfg.StorageSize == "" {
		cfg.StorageSize = "1Gi"
	}

	ssMap := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "StatefulSet",
		"metadata": map[string]interface{}{
			"name": cfg.Name + "-statefulset",
		},
		"spec": map[string]interface{}{
			"serviceName": cfg.ServiceName,
			"replicas":    cfg.Replicas,
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{"app": cfg.Name},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{"app": cfg.Name},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":  cfg.Name,
							"image": cfg.Image,
							"ports": []map[string]interface{}{
								{"containerPort": cfg.ContainerPort},
							},
							"volumeMounts": []map[string]interface{}{
								{
									"name":      "data",
									"mountPath": "/data",
								},
							},
						},
					},
				},
			},
			"volumeClaimTemplates": []map[string]interface{}{
				{
					"metadata": map[string]interface{}{
						"name": "data",
					},
					"spec": map[string]interface{}{
						"accessModes": []string{"ReadWriteOnce"},
						"resources": map[string]interface{}{
							"requests": map[string]string{
								"storage": cfg.StorageSize,
							},
						},
					},
				},
			},
		},
	}

	if cfg.Namespace != "" {
		ssMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(ssMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var statefulsetCmd = &cobra.Command{
	Use:   "statefulset",
	Short: "Generate a Kubernetes StatefulSet manifest",
	Long:  `Interactively create a StatefulSet with persistent volume claim templates for stateful workloads.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, image, replicasStr, portStr, serviceName, storageSize string

		survey.AskOne(&survey.Input{Message: "StatefulSet app name:", Default: "my-app"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Container image:", Default: "postgres:16-alpine"}, &image)
		survey.AskOne(&survey.Input{Message: "Replicas:", Default: "3"}, &replicasStr)
		survey.AskOne(&survey.Input{Message: "Container port:", Default: "5432"}, &portStr)
		survey.AskOne(&survey.Input{Message: "Headless service name:", Default: strings.TrimSpace(name) + "-service"}, &serviceName)
		survey.AskOne(&survey.Input{Message: "PVC storage size per pod:", Default: "10Gi"}, &storageSize)

		replicas, _ := strconv.Atoi(strings.TrimSpace(replicasStr))
		port, _ := strconv.Atoi(strings.TrimSpace(portStr))

		cfg := StatefulSetConfig{
			Name:          strings.TrimSpace(name),
			Namespace:     strings.TrimSpace(namespace),
			Image:         strings.TrimSpace(image),
			Replicas:      replicas,
			ContainerPort: port,
			ServiceName:   strings.TrimSpace(serviceName),
			StorageSize:   strings.TrimSpace(storageSize),
		}
		content := GenerateStatefulSet(cfg)
		writeK8sFile("statefulset.yaml", content)
	},
}

// ============================================================================
// DaemonSet
// ============================================================================

type DaemonSetConfig struct {
	Name          string
	Namespace     string
	Image         string
	ContainerPort int
	HostNetwork   bool
}

func GenerateDaemonSet(cfg DaemonSetConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-daemonset"
	}

	container := map[string]interface{}{
		"name":  cfg.Name,
		"image": cfg.Image,
	}
	if cfg.ContainerPort > 0 {
		container["ports"] = []map[string]interface{}{
			{"containerPort": cfg.ContainerPort},
		}
	}

	podSpec := map[string]interface{}{
		"containers": []map[string]interface{}{container},
	}
	if cfg.HostNetwork {
		podSpec["hostNetwork"] = true
	}

	dsMap := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
			"labels": map[string]string{
				"app": cfg.Name,
			},
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchLabels": map[string]string{"app": cfg.Name},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]string{"app": cfg.Name},
				},
				"spec": podSpec,
			},
		},
	}

	if cfg.Namespace != "" {
		dsMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(dsMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var daemonsetCmd = &cobra.Command{
	Use:   "daemonset",
	Short: "Generate a Kubernetes DaemonSet manifest",
	Long:  `Interactively create a DaemonSet that runs a pod on every node in the cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, image, portStr string
		var hostNet bool

		survey.AskOne(&survey.Input{Message: "DaemonSet name:", Default: "my-daemonset"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Container image:", Default: "fluentd:latest"}, &image)
		survey.AskOne(&survey.Input{Message: "Container port (0 for none):", Default: "0"}, &portStr)
		survey.AskOne(&survey.Confirm{Message: "Use host network?", Default: false}, &hostNet)

		port, _ := strconv.Atoi(strings.TrimSpace(portStr))

		cfg := DaemonSetConfig{
			Name:          strings.TrimSpace(name),
			Namespace:     strings.TrimSpace(namespace),
			Image:         strings.TrimSpace(image),
			ContainerPort: port,
			HostNetwork:   hostNet,
		}
		content := GenerateDaemonSet(cfg)
		writeK8sFile("daemonset.yaml", content)
	},
}

// ============================================================================
// Job
// ============================================================================

type JobConfig struct {
	Name           string
	Namespace      string
	Image          string
	Command        []string
	BackoffLimit   int
	Completions    int
	Parallelism    int
}

func GenerateJob(cfg JobConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-job"
	}
	if cfg.BackoffLimit < 0 {
		cfg.BackoffLimit = 4
	}

	container := map[string]interface{}{
		"name":  cfg.Name,
		"image": cfg.Image,
	}
	if len(cfg.Command) > 0 {
		container["command"] = cfg.Command
	}

	jobSpec := map[string]interface{}{
		"backoffLimit": cfg.BackoffLimit,
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"containers":    []map[string]interface{}{container},
				"restartPolicy": "Never",
			},
		},
	}

	if cfg.Completions > 0 {
		jobSpec["completions"] = cfg.Completions
	}
	if cfg.Parallelism > 0 {
		jobSpec["parallelism"] = cfg.Parallelism
	}

	jobMap := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"spec": jobSpec,
	}

	if cfg.Namespace != "" {
		jobMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	data, err := yaml.Marshal(jobMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Generate a Kubernetes Job manifest",
	Long:  `Interactively create a one-off Job with completions, parallelism, and backoff configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, image, cmdStr, backoffStr, compStr, parallelStr string

		survey.AskOne(&survey.Input{Message: "Job name:", Default: "my-job"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Container image:", Default: "busybox:latest"}, &image)
		survey.AskOne(&survey.Input{Message: "Command (space-separated):", Default: "/bin/sh -c echo done"}, &cmdStr)
		survey.AskOne(&survey.Input{Message: "Backoff limit:", Default: "4"}, &backoffStr)
		survey.AskOne(&survey.Input{Message: "Completions (0 for default):", Default: "1"}, &compStr)
		survey.AskOne(&survey.Input{Message: "Parallelism (0 for default):", Default: "1"}, &parallelStr)

		backoff, _ := strconv.Atoi(strings.TrimSpace(backoffStr))
		comp, _ := strconv.Atoi(strings.TrimSpace(compStr))
		parallel, _ := strconv.Atoi(strings.TrimSpace(parallelStr))

		var command []string
		if strings.TrimSpace(cmdStr) != "" {
			command = strings.Fields(cmdStr)
		}

		cfg := JobConfig{
			Name:         strings.TrimSpace(name),
			Namespace:    strings.TrimSpace(namespace),
			Image:        strings.TrimSpace(image),
			Command:      command,
			BackoffLimit: backoff,
			Completions:  comp,
			Parallelism:  parallel,
		}
		content := GenerateJob(cfg)
		writeK8sFile("job.yaml", content)
	},
}

// ============================================================================
// Namespace
// ============================================================================

func GenerateNamespace(name string, labels map[string]string) string {
	nsMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]interface{}{
			"name": name,
		},
	}
	if len(labels) > 0 {
		nsMap["metadata"].(map[string]interface{})["labels"] = labels
	}

	data, err := yaml.Marshal(nsMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var namespaceCmd = &cobra.Command{
	Use:   "namespace",
	Short: "Generate a Kubernetes Namespace manifest",
	Long:  `Create a Namespace manifest with optional labels.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, labelsStr string
		survey.AskOne(&survey.Input{Message: "Namespace name:", Default: "my-namespace"}, &name)
		survey.AskOne(&survey.Input{Message: "Labels (comma-separated key=value, blank for none):"}, &labelsStr)

		labels := make(map[string]string)
		if strings.TrimSpace(labelsStr) != "" {
			for _, pair := range strings.Split(labelsStr, ",") {
				parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
				if len(parts) == 2 {
					labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		content := GenerateNamespace(strings.TrimSpace(name), labels)
		writeK8sFile("namespace.yaml", content)
	},
}

// ============================================================================
// ServiceAccount + RBAC
// ============================================================================

type RBACConfig struct {
	Name            string
	Namespace       string
	ClusterWide     bool
	Rules           []RBACRule
}

type RBACRule struct {
	APIGroups []string
	Resources []string
	Verbs     []string
}

func GenerateRBAC(cfg RBACConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app"
	}

	var sb strings.Builder

	// ServiceAccount
	saMap := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ServiceAccount",
		"metadata": map[string]interface{}{
			"name": cfg.Name + "-sa",
		},
	}
	if cfg.Namespace != "" {
		saMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}
	saData, _ := yaml.Marshal(saMap)
	sb.Write(saData)
	sb.WriteString("---\n")

	// Role or ClusterRole
	roleKind := "Role"
	roleBindingKind := "RoleBinding"
	roleAPIVersion := "rbac.authorization.k8s.io/v1"
	if cfg.ClusterWide {
		roleKind = "ClusterRole"
		roleBindingKind = "ClusterRoleBinding"
	}

	var rules []map[string]interface{}
	for _, r := range cfg.Rules {
		rules = append(rules, map[string]interface{}{
			"apiGroups": r.APIGroups,
			"resources": r.Resources,
			"verbs":     r.Verbs,
		})
	}

	if len(rules) == 0 {
		rules = []map[string]interface{}{
			{
				"apiGroups": []string{""},
				"resources": []string{"pods", "services", "configmaps"},
				"verbs":     []string{"get", "list", "watch"},
			},
		}
	}

	roleMap := map[string]interface{}{
		"apiVersion": roleAPIVersion,
		"kind":       roleKind,
		"metadata": map[string]interface{}{
			"name": cfg.Name + "-role",
		},
		"rules": rules,
	}
	if !cfg.ClusterWide && cfg.Namespace != "" {
		roleMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}
	roleData, _ := yaml.Marshal(roleMap)
	sb.Write(roleData)
	sb.WriteString("---\n")

	// RoleBinding or ClusterRoleBinding
	bindingMap := map[string]interface{}{
		"apiVersion": roleAPIVersion,
		"kind":       roleBindingKind,
		"metadata": map[string]interface{}{
			"name": cfg.Name + "-rolebinding",
		},
		"subjects": []map[string]interface{}{
			{
				"kind":      "ServiceAccount",
				"name":      cfg.Name + "-sa",
				"namespace": cfg.Namespace,
			},
		},
		"roleRef": map[string]interface{}{
			"apiGroup": "rbac.authorization.k8s.io",
			"kind":     roleKind,
			"name":     cfg.Name + "-role",
		},
	}
	if !cfg.ClusterWide && cfg.Namespace != "" {
		bindingMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}
	bindingData, _ := yaml.Marshal(bindingMap)
	sb.Write(bindingData)

	return sb.String()
}

var rbacCmd = &cobra.Command{
	Use:   "rbac",
	Short: "Generate ServiceAccount, Role, and RoleBinding manifests",
	Long:  `Interactively create RBAC resources: ServiceAccount, Role/ClusterRole, and RoleBinding/ClusterRoleBinding.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace string
		var clusterWide bool

		survey.AskOne(&survey.Input{Message: "App/ServiceAccount base name:", Default: "my-app"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace:", Default: "default"}, &namespace)
		survey.AskOne(&survey.Confirm{Message: "Cluster-wide permissions (ClusterRole)?", Default: false}, &clusterWide)

		var resourcesStr, verbsStr string
		survey.AskOne(&survey.Input{
			Message: "Resources (comma-separated, e.g. pods,services,configmaps):",
			Default: "pods,services,configmaps",
		}, &resourcesStr)
		survey.AskOne(&survey.Input{
			Message: "Verbs (comma-separated, e.g. get,list,watch,create,update,delete):",
			Default: "get,list,watch",
		}, &verbsStr)

		resources := splitTrimmed(resourcesStr, ",")
		verbs := splitTrimmed(verbsStr, ",")

		cfg := RBACConfig{
			Name:        strings.TrimSpace(name),
			Namespace:   strings.TrimSpace(namespace),
			ClusterWide: clusterWide,
			Rules: []RBACRule{
				{APIGroups: []string{""}, Resources: resources, Verbs: verbs},
			},
		}
		content := GenerateRBAC(cfg)
		writeK8sFile("rbac.yaml", content)
	},
}

// ============================================================================
// NetworkPolicy
// ============================================================================

type NetworkPolicyConfig struct {
	Name       string
	Namespace  string
	PodLabel   string
	IngressFrom string // "all", "same-namespace", "specific-label"
	EgressTo    string // "all", "deny-all"
	Port       int
}

func GenerateNetworkPolicy(cfg NetworkPolicyConfig) string {
	if cfg.Name == "" {
		cfg.Name = "my-app-netpol"
	}

	npMap := map[string]interface{}{
		"apiVersion": "networking.k8s.io/v1",
		"kind":       "NetworkPolicy",
		"metadata": map[string]interface{}{
			"name": cfg.Name,
		},
		"spec": map[string]interface{}{
			"podSelector": map[string]interface{}{
				"matchLabels": map[string]string{"app": cfg.PodLabel},
			},
			"policyTypes": []string{"Ingress", "Egress"},
		},
	}

	if cfg.Namespace != "" {
		npMap["metadata"].(map[string]interface{})["namespace"] = cfg.Namespace
	}

	spec := npMap["spec"].(map[string]interface{})

	// Ingress rules
	switch cfg.IngressFrom {
	case "all":
		ingressRule := map[string]interface{}{}
		if cfg.Port > 0 {
			ingressRule["ports"] = []map[string]interface{}{
				{"protocol": "TCP", "port": cfg.Port},
			}
		}
		spec["ingress"] = []map[string]interface{}{ingressRule}
	case "same-namespace":
		ingressRule := map[string]interface{}{
			"from": []map[string]interface{}{
				{"podSelector": map[string]interface{}{}},
			},
		}
		if cfg.Port > 0 {
			ingressRule["ports"] = []map[string]interface{}{
				{"protocol": "TCP", "port": cfg.Port},
			}
		}
		spec["ingress"] = []map[string]interface{}{ingressRule}
	case "deny-all":
		spec["ingress"] = []map[string]interface{}{}
	}

	// Egress rules
	switch cfg.EgressTo {
	case "all":
		spec["egress"] = []map[string]interface{}{{}}
	case "deny-all":
		spec["egress"] = []map[string]interface{}{}
	}

	data, err := yaml.Marshal(npMap)
	if err != nil {
		return ""
	}
	return string(data)
}

var netpolCmd = &cobra.Command{
	Use:   "netpol",
	Short: "Generate a Kubernetes NetworkPolicy manifest",
	Long:  `Interactively create a NetworkPolicy to control ingress and egress traffic for pods.`,
	Run: func(cmd *cobra.Command, args []string) {
		var name, namespace, podLabel, ingressFrom, egressTo, portStr string

		survey.AskOne(&survey.Input{Message: "NetworkPolicy name:", Default: "my-app-netpol"}, &name)
		survey.AskOne(&survey.Input{Message: "Namespace (blank for default):"}, &namespace)
		survey.AskOne(&survey.Input{Message: "Target pod label (app=?):", Default: "my-app"}, &podLabel)
		survey.AskOne(&survey.Select{
			Message: "Allow ingress from:",
			Options: []string{"all", "same-namespace", "deny-all"},
			Default: "same-namespace",
		}, &ingressFrom)
		survey.AskOne(&survey.Select{
			Message: "Allow egress to:",
			Options: []string{"all", "deny-all"},
			Default: "all",
		}, &egressTo)
		survey.AskOne(&survey.Input{Message: "Ingress port (0 for any):", Default: "0"}, &portStr)

		port, _ := strconv.Atoi(strings.TrimSpace(portStr))

		cfg := NetworkPolicyConfig{
			Name:        strings.TrimSpace(name),
			Namespace:   strings.TrimSpace(namespace),
			PodLabel:    strings.TrimSpace(podLabel),
			IngressFrom: ingressFrom,
			EgressTo:    egressTo,
			Port:        port,
		}
		content := GenerateNetworkPolicy(cfg)
		writeK8sFile("networkpolicy.yaml", content)
	},
}

// ============================================================================
// Shared helpers
// ============================================================================

func writeK8sFile(filename, content string) {
	k8sDir := "k8s"
	if err := os.MkdirAll(k8sDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", k8sDir, err)
		return
	}

	outPath := filepath.Join(k8sDir, filename)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("✓ Generated %s\n", filepath.ToSlash(outPath))
}

func splitTrimmed(s string, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ============================================================================
// Command registration
// ============================================================================

func init() {
	clusterCmd.AddCommand(ingressCmd)
	clusterCmd.AddCommand(configmapCmd)
	clusterCmd.AddCommand(secretCmd)
	clusterCmd.AddCommand(hpaCmd)
	clusterCmd.AddCommand(cronjobCmd)
	clusterCmd.AddCommand(pvcCmd)
	clusterCmd.AddCommand(statefulsetCmd)
	clusterCmd.AddCommand(daemonsetCmd)
	clusterCmd.AddCommand(jobCmd)
	clusterCmd.AddCommand(namespaceCmd)
	clusterCmd.AddCommand(rbacCmd)
	clusterCmd.AddCommand(netpolCmd)
}
