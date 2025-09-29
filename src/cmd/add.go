// cmd/add.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	// keyName   string
	keyFile   string
	overwrite bool
)

var addCmd = &cobra.Command{
	Use:   "add [host]",
	Short: "Add SSH public key for a host",
	Long: `Add an SSH public key to the key management system.
You can specify either a key file or let the tool generate a new key pair.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]

		if keyFile != "" {
			if err := addKeyFromFile(host, keyFile); err != nil {
				fmt.Printf("Error adding key from file: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := generateAndAddKey(host, keyName); err != nil {
				fmt.Printf("Error generating key: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Successfully added key for host: %s\n", host)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&keyName, "name", "n", "default", "name for the key")
	addCmd.Flags().StringVarP(&keyFile, "file", "f", "", "public key file to add")
	addCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite existing key")
}

func addKeyFromFile(host, keyFile string) error {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// 这里实现添加密钥的逻辑
	fmt.Printf("Adding key from %s for host %s\n", keyFile, host)
	fmt.Printf("Key content: %s\n", string(data[:50])+"...")

	return nil
}

func generateAndAddKey(host, keyName string) error {
	// 这里实现生成新密钥对的逻辑
	fmt.Printf("Generating new key pair for host %s with name %s\n", host, keyName)

	privateKeyPath := filepath.Join(configDir, host, keyName)
	publicKeyPath := privateKeyPath + ".pub"

	fmt.Printf("Private key will be saved to: %s\n", privateKeyPath)
	fmt.Printf("Public key will be saved to: %s\n", publicKeyPath)

	// 模拟密钥生成
	fmt.Println("Key pair generated successfully!")

	return nil
}
