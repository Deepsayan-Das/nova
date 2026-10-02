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

var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Scaffold a Helm chart for Kubernetes deployments",
	Long: `helm generates a complete Helm chart directory structure including:
  - Chart.yaml (chart metadata)
  - values.yaml (configurable values)
  - templates/deployment.yaml
  - templates/service.yaml
  - templates/ingress.yaml
  - templates/hpa.yaml
  - templates/serviceaccount.yaml
  - templates/_helpers.tpl (template helpers)
  - templates/NOTES.txt (post-install notes)
  - .helmignore

This provides a production-ready Helm chart that can be customized and deployed.`,
	Run: runHelm,
}

func runHelm(cmd *cobra.Command, args []string) {
	var chartName, description, appVersion, chartVersion string
	var port, replicas string

	survey.AskOne(&survey.Input{Message: "Chart name:", Default: "my-app"}, &chartName)
	survey.AskOne(&survey.Input{Message: "Description:", Default: "A Helm chart for Kubernetes"}, &description)
	survey.AskOne(&survey.Input{Message: "Chart version:", Default: "0.1.0"}, &chartVersion)
	survey.AskOne(&survey.Input{Message: "Application version:", Default: "1.0.0"}, &appVersion)
	survey.AskOne(&survey.Input{Message: "Default container port:", Default: "8080"}, &port)
	survey.AskOne(&survey.Input{Message: "Default replica count:", Default: "2"}, &replicas)

	chartName = strings.TrimSpace(chartName)
	chartDir := filepath.Join("helm", chartName)
	templatesDir := filepath.Join(chartDir, "templates")

	for _, dir := range []string{chartDir, templatesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	generated := 0

	// Chart.yaml
	chartYaml := fmt.Sprintf(`apiVersion: v2
name: %s
description: %s
type: application
version: %s
appVersion: "%s"
maintainers:
  - name: Your Name
    email: your-email@example.com
`, chartName, description, chartVersion, appVersion)
	writeHelmFile(chartDir, "Chart.yaml", chartYaml)
	generated++

	// values.yaml
	valuesYaml := fmt.Sprintf(`# Default values for %s
replicaCount: %s

image:
  repository: nginx
  pullPolicy: IfNotPresent
  tag: ""

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations: {}
podSecurityContext: {}
securityContext: {}

service:
  type: ClusterIP
  port: %s

ingress:
  enabled: false
  className: "nginx"
  annotations: {}
  hosts:
    - host: %s.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 250m
    memory: 64Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80

nodeSelector: {}
tolerations: []
affinity: {}

env: []

configMap:
  enabled: false
  data: {}
`, chartName, replicas, port, chartName)
	writeHelmFile(chartDir, "values.yaml", valuesYaml)
	generated++

	// templates/_helpers.tpl
	helpersTpl := fmt.Sprintf(`{{/*
Expand the name of the chart.
*/}}
{{- define "%s.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "%s.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%%s-%%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "%s.chart" -}}
{{- printf "%%s-%%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "%s.labels" -}}
helm.sh/chart: {{ include "%s.chart" . }}
{{ include "%s.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "%s.selectorLabels" -}}
app.kubernetes.io/name: {{ include "%s.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "%s.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "%s.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
`, chartName, chartName, chartName, chartName, chartName, chartName, chartName, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "_helpers.tpl", helpersTpl)
	generated++

	// templates/deployment.yaml
	deploymentTpl := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "%s.fullname" . }}
  labels:
    {{- include "%s.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "%s.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "%s.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "%s.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.service.port }}
              protocol: TCP
          {{- with .Values.env }}
          env:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          livenessProbe:
            httpGet:
              path: /
              port: http
          readinessProbe:
            httpGet:
              path: /
              port: http
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
`, chartName, chartName, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "deployment.yaml", deploymentTpl)
	generated++

	// templates/service.yaml
	serviceTpl := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: {{ include "%s.fullname" . }}
  labels:
    {{- include "%s.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "%s.selectorLabels" . | nindent 4 }}
`, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "service.yaml", serviceTpl)
	generated++

	// templates/ingress.yaml
	ingressTpl := fmt.Sprintf(`{{- if .Values.ingress.enabled -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "%s.fullname" . }}
  labels:
    {{- include "%s.labels" . | nindent 4 }}
  {{- with .Values.ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if .Values.ingress.className }}
  ingressClassName: {{ .Values.ingress.className }}
  {{- end }}
  {{- if .Values.ingress.tls }}
  tls:
    {{- range .Values.ingress.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ . | quote }}
        {{- end }}
      secretName: {{ .secretName }}
    {{- end }}
  {{- end }}
  rules:
    {{- range .Values.ingress.hosts }}
    - host: {{ .host | quote }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            pathType: {{ .pathType }}
            backend:
              service:
                name: {{ include "%s.fullname" $ }}
                port:
                  number: {{ $.Values.service.port }}
          {{- end }}
    {{- end }}
{{- end }}
`, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "ingress.yaml", ingressTpl)
	generated++

	// templates/hpa.yaml
	hpaTpl := fmt.Sprintf(`{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "%s.fullname" . }}
  labels:
    {{- include "%s.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "%s.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
    {{- if .Values.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
    {{- end }}
    {{- if .Values.autoscaling.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
    {{- end }}
{{- end }}
`, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "hpa.yaml", hpaTpl)
	generated++

	// templates/serviceaccount.yaml
	saTpl := fmt.Sprintf(`{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "%s.serviceAccountName" . }}
  labels:
    {{- include "%s.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}
`, chartName, chartName)
	writeHelmFile(templatesDir, "serviceaccount.yaml", saTpl)
	generated++

	// templates/NOTES.txt
	notesTxt := fmt.Sprintf(`1. Get the application URL by running these commands:
{{- if .Values.ingress.enabled }}
  http{{ if $.Values.ingress.tls }}s{{ end }}://{{ (index .Values.ingress.hosts 0).host }}
{{- else if contains "NodePort" .Values.service.type }}
  export NODE_PORT=$(kubectl get --namespace {{ .Release.Namespace }} -o jsonpath="{.spec.ports[0].nodePort}" services {{ include "%s.fullname" . }})
  export NODE_IP=$(kubectl get nodes --namespace {{ .Release.Namespace }} -o jsonpath="{.items[0].status.addresses[0].address}")
  echo http://$NODE_IP:$NODE_PORT
{{- else if contains "LoadBalancer" .Values.service.type }}
  NOTE: It may take a few minutes for the LoadBalancer IP to be available.
  kubectl get --namespace {{ .Release.Namespace }} svc {{ include "%s.fullname" . }} -w
{{- else if contains "ClusterIP" .Values.service.type }}
  kubectl --namespace {{ .Release.Namespace }} port-forward svc/{{ include "%s.fullname" . }} {{ .Values.service.port }}:{{ .Values.service.port }}
  echo "Visit http://127.0.0.1:{{ .Values.service.port }}"
{{- end }}
`, chartName, chartName, chartName)
	writeHelmFile(templatesDir, "NOTES.txt", notesTxt)
	generated++

	// .helmignore
	helmignore := `# Patterns to ignore when building packages.
.git
.gitignore
.DS_Store
*.swp
*.bak
*.tmp
*.orig
*~
.project
.idea/
*.tmproj
.vscode/
`
	writeHelmFile(chartDir, ".helmignore", helmignore)
	generated++

	fmt.Printf("\n✓ Generated Helm chart with %d files in %s/\n", generated, filepath.ToSlash(chartDir))
	fmt.Printf("  Install with: helm install %s %s/\n", chartName, filepath.ToSlash(chartDir))
	fmt.Printf("  Package with: helm package %s/\n", filepath.ToSlash(chartDir))
}

func writeHelmFile(dir, filename, content string) {
	outPath := filepath.Join(dir, filename)
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing %s: %v\n", outPath, err)
		return
	}
	fmt.Printf("  ✓ %s\n", filepath.ToSlash(outPath))
}

func init() {
	clusterCmd.AddCommand(helmCmd)
}
