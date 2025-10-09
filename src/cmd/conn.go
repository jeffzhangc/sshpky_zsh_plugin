package cmd

import (
	"fmt"
	"os"
	"sshpky/pkg/config"
	"sshpky/pkg/sshrunner"
	"strings"

	"github.com/spf13/cobra"
)

var connArgs config.SshConfigItem

var connectCmd = &cobra.Command{
	Use:   "conn [user@]host",
	Short: "Connet to a host using managed SSH keys",
	Long: `Connet to a remote host using one of the managed SSH keys.
This command will automatically select and use the appropriate key.

Examples:
  # Connect with default user (current user)
  sshpky conn example.com

  # Connect with specific user
  sshpky conn user@example.com

  # Connect with specific identity file
  sshpky conn -i ~/.ssh/custom_key user@example.com`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: connValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := parseDestination(args[0], &connArgs); err != nil {
			fmt.Printf("Error parsing destination: %v\n", err)
			os.Exit(1)
		}

		runConn(connArgs, args)
	},
}

// connValidArgs 为 conn 命令提供自动补全建议
func connValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// 如果已经输入了参数，不再提供补全
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 获取 group 参数值
	group, _ := cmd.Flags().GetString("group")

	// 加载配置
	manager := config.NewSSHConfigManager("")
	var configs []*config.SshConfigItem
	var err error

	if group == "" {
		cfg := config.GetConfig()
		group = cfg.Use
	}

	if group != "" {
		// 如果指定了 group，只显示该 group 下的 hosts
		configs, err = manager.GetConfigsByGroup(group)
		// } else {
		// 	// 如果没有指定 group，使用默认 group
		// 	cfg := config.GetConfig()
		// 	if cfg.Use != "" {
		// 		configs, err = manager.GetConfigsByGroup(cfg.Use)
		// 	} else {
		// 		// 如果没有默认 group，显示所有 hosts
		// 		configs, err = manager.ReadConfig()
		// }
	}

	if err != nil {
		// 出错时返回空列表
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// 构建补全建议列表
	var suggestions []string
	for _, config := range configs {
		// 添加 host 作为补全建议
		suggestions = append(suggestions, config.Host)

		// // 也可以添加 user@host 格式的补全建议
		// if config.User != "" {
		// 	suggestions = append(suggestions, fmt.Sprintf("%s@%s", config.User, config.Host))
		// }
	}

	// 过滤以 toComplete 开头的建议
	var filtered []string
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, toComplete) {
			filtered = append(filtered, suggestion)
		}
	}

	// 如果没有匹配的，返回所有建议
	if len(filtered) == 0 {
		filtered = suggestions
	}

	return filtered, cobra.ShellCompDirectiveNoFileComp
}
func runConn(connArgs config.SshConfigItem, args []string) {
	// 构建 SSH 命令
	sshCmd := buildSSHCommand(connArgs)
	fmt.Println("Executing:", sshCmd)

	// 这里可以添加实际的 SSH 连接逻辑
	// exec.Command("ssh", sshArgs...)
	err := sshrunner.RunSSH(sshCmd, connArgs, args)
	if err != nil {
		// panic(err)
		fmt.Println("error", err.Error())
	}
	fmt.Println("done..")
}

// parseDestination 解析目标地址，支持 user@host 格式
func parseDestination(destination string, connArgs *config.SshConfigItem) error {
	if strings.Contains(destination, "@") {
		parts := strings.SplitN(destination, "@", 2)
		connArgs.User = parts[0]
		connArgs.Host = parts[1]
	} else {
		connArgs.Host = destination
	}
	return nil
}

// buildSSHCommand 构建 SSH 命令参数
func buildSSHCommand(connArgs config.SshConfigItem) string {
	var args []string

	// 添加端口参数
	if connArgs.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", connArgs.Port))
	}
	// 添加身份文件参数（-i）
	if connArgs.IdentityFile != "" {
		args = append(args, "-i", connArgs.IdentityFile)
	}

	// 添加目标地址
	var target string
	if connArgs.User == "" {
		target = connArgs.Host
	} else {
		target = fmt.Sprintf("%s@%s", connArgs.User, connArgs.Host)
	}
	args = append(args, target)

	// 构建命令字符串
	cmdStr := "ssh"
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			cmdStr += " \"" + arg + "\""
		} else {
			cmdStr += " " + arg
		}
	}

	return cmdStr
}

// groupFlagValidArgs 为 group 标志提供自动补全建议
func groupFlagValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return groupUseValidArgs(cmd, args, toComplete)
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// 用户参数 - 如果同时在 destination 和 -u 中指定，以 -u 为准
	connectCmd.Flags().StringVarP(&connArgs.User, "user", "u", "", "username for SSH connection")

	// 端口参数
	connectCmd.Flags().IntVarP(&connArgs.Port, "port", "p", 22, "SSH port")

	// 身份文件参数（标准 SSH 的 -i 参数）
	connectCmd.Flags().StringVarP(&connArgs.IdentityFile, "identity", "i", "", "identity file (private key) for public key authentication")

	// group 信息
	connectCmd.Flags().StringVarP(&connArgs.Group, "group", "g", "", "group for this ssh")

	connectCmd.Flags().StringVarP(&connArgs.HostName, "hostname", "", "", "hostname for this ssh")

	// 为 group 标志添加补全
	connectCmd.RegisterFlagCompletionFunc("group", groupFlagValidArgs)
}
