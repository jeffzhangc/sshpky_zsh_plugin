// cmd/remove.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [host]",
	Short: "Remove SSH public key for a host",
	Long:  `Remove a previously added SSH public key for the specified host.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]

		if err := removeKey(host, keyName); err != nil {
			fmt.Printf("Error removing key: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully removed key for host: %s\n", host)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringVarP(&keyName, "name", "n", "default", "name of the key to remove")
}

func removeKey(host, keyName string) error {
	// 这里实现移除密钥的逻辑
	fmt.Printf("Removing key %s for host %s\n", keyName, host)
	return nil
}
