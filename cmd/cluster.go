/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage and generate Kubernetes cluster resources",
	Long: `Cluster commands allow you to generate Kubernetes manifests, Helm charts,
Kustomize overlays, and manage cluster deployments via kubectl.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
