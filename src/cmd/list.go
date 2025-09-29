// cmd/list.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [host]",
	Short: "List all managed SSH public keys",
	Long: `List all managed SSH public keys.
If a host is specified, only show keys for that host.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var host string
		if len(args) > 0 {
			host = args[0]
		}

		if err := listKeys(host); err != nil {
			fmt.Printf("Error listing keys: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listKeys(host string) error {
	// 这里实现列出密钥的逻辑
	if host == "" {
		fmt.Println("Managed SSH keys:")
		fmt.Println("-----------------")

		// 模拟列出所有主机的密钥
		hosts := []string{"server1", "server2", "server3"}
		for _, h := range hosts {
			fmt.Printf("Host: %s\n", h)
			fmt.Printf("  - default (RSA 2048)\n")
			fmt.Printf("  - backup (Ed25519)\n")
			fmt.Println()
		}
	} else {
		fmt.Printf("Keys for host %s:\n", host)
		fmt.Println("-----------------")
		fmt.Printf("  - default (RSA 2048)\n")
		fmt.Printf("    Location: %s\n", filepath.Join(configDir, host, "default"))
		fmt.Printf("  - backup (Ed25519)\n")
		fmt.Printf("    Location: %s\n", filepath.Join(configDir, host, "backup"))
	}

	return nil
}
