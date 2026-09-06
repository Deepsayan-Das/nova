package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Nova CLI",
	Long:  `All software has versions. This is Nova's CLI version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nova version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
