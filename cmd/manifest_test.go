package cmd

import (
	"strings"
	"testing"
)

func TestGenerateK8sDeployment(t *testing.T) {
	cfg := K8sConfig{
		AppName:       "nova-web",
		Image:         "deeps/nova-web:v1.0",
		Replicas:      3,
		ContainerPort: 8080,
		ServiceType:   "LoadBalancer",
	}

	got := GenerateK8sDeployment(cfg)

	if !strings.Contains(got, "kind: Deployment") {
		t.Errorf("Expected kind: Deployment, got:\n%s", got)
	}
	if !strings.Contains(got, "name: nova-web-deployment") {
		t.Errorf("Expected deployment name nova-web-deployment, got:\n%s", got)
	}
	if !strings.Contains(got, "replicas: 3") {
		t.Errorf("Expected replicas 3, got:\n%s", got)
	}
	if !strings.Contains(got, "image: deeps/nova-web:v1.0") {
		t.Errorf("Expected image deeps/nova-web:v1.0, got:\n%s", got)
	}
	if !strings.Contains(got, "containerPort: 8080") {
		t.Errorf("Expected containerPort 8080, got:\n%s", got)
	}
}

func TestGenerateK8sService(t *testing.T) {
	cfg := K8sConfig{
		AppName:       "nova-web",
		Image:         "deeps/nova-web:v1.0",
		Replicas:      3,
		ContainerPort: 8080,
		ServiceType:   "LoadBalancer",
	}

	got := GenerateK8sService(cfg)

	if !strings.Contains(got, "kind: Service") {
		t.Errorf("Expected kind: Service, got:\n%s", got)
	}
	if !strings.Contains(got, "name: nova-web-service") {
		t.Errorf("Expected service name nova-web-service, got:\n%s", got)
	}
	if !strings.Contains(got, "type: LoadBalancer") {
		t.Errorf("Expected type LoadBalancer, got:\n%s", got)
	}
	if !strings.Contains(got, "port: 8080") {
		t.Errorf("Expected port 8080, got:\n%s", got)
	}
}

func TestGenerateK8sManifest(t *testing.T) {
	cfg := K8sConfig{
		AppName:       "api-service",
		Image:         "api:latest",
		Replicas:      1,
		ContainerPort: 3000,
		ServiceType:   "ClusterIP",
	}

	dep, svc := GenerateK8sManifest(cfg)

	if !strings.Contains(dep, "kind: Deployment") || !strings.Contains(dep, "api-service") {
		t.Errorf("GenerateK8sManifest deployment invalid:\n%s", dep)
	}
	if !strings.Contains(svc, "kind: Service") || !strings.Contains(svc, "ClusterIP") {
		t.Errorf("GenerateK8sManifest service invalid:\n%s", svc)
	}
}
