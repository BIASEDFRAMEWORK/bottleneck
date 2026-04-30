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
		if bottleneckRootExists(".") {
			fmt.Println("System initialized")
			return
		}

		fmt.Println("System incomplete")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func bottleneckRootExists(basePath string) bool {
	info, err := os.Stat(filepath.Join(basePath, "bottleneck"))
	return err == nil && info.IsDir()
}
