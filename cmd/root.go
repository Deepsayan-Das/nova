/*
Copyright © 2026 Deepsayan-Das <deepsayandas274@gmail.com>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "nova",
	Version: Version,
	Short:   "A CLI for GalactOS Operating System",
	Long:  `Nova is the root command line interface for the GalactOS Operating System. It provides various subcommands to manage and interact with the system, including configuration, monitoring, and maintenance tasks.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nova.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}
