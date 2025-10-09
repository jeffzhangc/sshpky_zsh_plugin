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
	Use:   "sshpky [flags] [HOST]",
	Short: "SSH Public Key Management Tool",
	Long: `A comprehensive tool for managing SSH public keys and connections.

Simplify your SSH key management and server connections with automatic
key selection, group management, and secure credential storage.

If a HOST argument is provided, it will automatically connect to that host
using the connection command. Otherwise, you can use subcommands for
specific operations.

Examples:
  # Quick connect to a host (default behavior)
  sshpky 10.1.102.32
  sshpky user@example.com

  # Connect with specific options
  sshpky -u admin -p 2222 example.com
  sshpky -g production webserver

  # Use subcommands for other operations
  sshpky mg list                    # List all groups
  sshpky mg use production         # Switch to production group
  sshpky conn -g staging appserver  # Connect to specific group's host`,
	Args: cobra.ArbitraryArgs, // 允许任意参数
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 确保配置目录存在
		if err := ensureConfigDir(); err != nil {
			fmt.Printf("Error creating config directory: %v\n", err)
			os.Exit(1)
		}
	},
	// 如果没有子命令，默认执行 conn 命令
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有参数，显示帮助信息
		if len(args) == 0 {
			cmd.Help()
			return
		}

		// 否则执行 conn 命令逻辑
		runDefaultConn(args)
	},
	// 为 root 命令添加参数补全
	ValidArgsFunction: rootValidArgs,
}

// rootValidArgs 为 root 命令提供自动补全建议
func rootValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 如果已经有参数，不再提供补全
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 获取所有可用的 hosts 作为补全建议
	return connValidArgs(cmd, args, toComplete)
}

// runDefaultConn 执行默认的 conn 命令逻辑

func runDefaultConn(args []string) {
	// 解析目标地址
	if err := parseDestination(args[0], &connArgs); err != nil {
		fmt.Printf("Error parsing destination: %v\n", err)
		os.Exit(1)
	}

	// 运行连接
	runConn(connArgs, args)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultConfigDir := filepath.Join(homeDir, ".sshpky")

	rootCmd.PersistentFlags().StringVarP(&configDir, "config-dir", "c", defaultConfigDir, "config directory")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "config.yaml", "config file name")

	// 用户参数 - 如果同时在 destination 和 -u 中指定，以 -u 为准
	rootCmd.Flags().StringVarP(&connArgs.User, "user", "u", "", "username for SSH connection")

	// 端口参数
	rootCmd.Flags().IntVarP(&connArgs.Port, "port", "p", 22, "SSH port")

	// 身份文件参数（标准 SSH 的 -i 参数）
	rootCmd.Flags().StringVarP(&connArgs.IdentityFile, "identity", "i", "", "identity file (private key) for public key authentication")

	// group 信息
	rootCmd.Flags().StringVarP(&connArgs.Group, "group", "g", "", "group for this ssh")

	rootCmd.Flags().StringVarP(&connArgs.HostName, "hostname", "", "", "hostname for this ssh")

	// 添加 completion 命令
	rootCmd.AddCommand(completionCmd)
}

func ensureConfigDir() error {
	return os.MkdirAll(configDir, 0700)
}

// 创建自定义的 completion 命令
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for sshpky.

This command generates completion scripts for various shells. After generating,
you need to source the file or add it to your shell's startup configuration.

Installation Instructions:

Bash (Linux/macOS):
  # Generate and save the completion script
  sshpky completion bash > ~/.sshpky-completion.bash
  
  # Add to your bashrc
  echo "source ~/.sshpky-completion.bash" >> ~/.bashrc
  
  # Reload your current shell
  source ~/.bashrc

  # Alternative: system-wide installation (Linux)
  sudo sshpky completion bash > /etc/bash_completion.d/sshpky

Bash (macOS with Homebrew):
  # If installed via Homebrew, completion might be automatically installed
  brew install bash-completion

Zsh:
  # Generate and save the completion script
  sshpky completion zsh > ~/.sshpky-completion.zsh
  
  # Add to your zshrc
  echo "source ~/.sshpky-completion.zsh" >> ~/.zshrc
  
  # Reload your current shell
  source ~/.zshrc

  # Alternative: use the function path
  sshpky completion zsh > "${fpath[1]}/_sshpky"

Fish:
  # Generate and save the completion script
  sshpky completion fish > ~/.config/fish/completions/sshpky.fish
  
  # Reload your current shell
  exec fish

PowerShell:
  # Generate and execute immediately
  sshpky completion powershell | Out-String | Invoke-Expression
  
  # To persist across sessions, add to your profile
  sshpky completion powershell > $PROFILE
  
  # Or create a separate file and source it
  sshpky completion powershell > ~/.sshpky-completion.ps1
  Add-Content $PROFILE "~/.sshpky-completion.ps1"

After installation, restart your shell or source the configuration file.
You can then use tab completion for hosts, groups, and commands.`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
