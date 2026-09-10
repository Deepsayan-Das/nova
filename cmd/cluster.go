/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage and generate Kubernetes cluster resources",
	Long:  `Cluster commands allow you to generate Kubernetes manifests and manage cluster configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cluster called. Use 'nova cluster manifest' to generate Kubernetes manifests.")
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
