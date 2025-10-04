package cmd

// cmd/root.go

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	configDir  string
	configFile string
)

var rootCmd = &cobra.Command{
	Use:   "sshpky",
	Short: "SSH Public Key Management Tool",
	Long: `A comprehensive tool for managing SSH public keys and connections.
Simplify your SSH key management and server connections.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 确保配置目录存在
		if err := ensureConfigDir(); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultConfigDir := filepath.Join(homeDir, ".sshpky")

	rootCmd.PersistentFlags().StringVarP(&configDir, "config-dir", "c", defaultConfigDir, "config directory")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "config.yaml", "config file name")
}

func ensureConfigDir() error {
	return os.MkdirAll(configDir, 0700)
}
