package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the current system status",
	Run: func(cmd *cobra.Command, args []string) {
		if biasedRootExists(".") {
			fmt.Println("System initialized")
			return
		}

		fmt.Println("System incomplete")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func biasedRootExists(basePath string) bool {
	info, err := os.Stat(filepath.Join(basePath, "biased"))
	return err == nil && info.IsDir()
}
