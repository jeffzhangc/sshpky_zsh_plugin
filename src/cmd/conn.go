package cmd

import (
	"fmt"
	"os"
	"sshpky/pkg/sshrunner"
	"strings"

	"github.com/spf13/cobra"
)

type ConnArgs struct {
	User     string
	Port     int
	Identity string // -i 参数，指定密钥文件
	Host     string
}

var connArgs ConnArgs

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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := parseDestination(args[0]); err != nil {
			fmt.Printf("Error parsing destination: %v\n", err)
			os.Exit(1)
		}

		runConn(connArgs, args)
	},
}

func runConn(connArgs ConnArgs, args []string) {
	// 构建 SSH 命令
	sshCmd := buildSSHCommand(connArgs)
	fmt.Println("Executing:", sshCmd)

	// 这里可以添加实际的 SSH 连接逻辑
	// exec.Command("ssh", sshArgs...)
	err := sshrunner.RunSSH(sshCmd, connArgs.User, connArgs.Host, connArgs.Port, args)
	if err != nil {
		// panic(err)
		fmt.Println("error", err.Error())
	}
	fmt.Println("done..")
}

// parseDestination 解析目标地址，支持 user@host 格式
func parseDestination(destination string) error {
	if strings.Contains(destination, "@") {
		parts := strings.SplitN(destination, "@", 2)
		connArgs.User = parts[0]
		connArgs.Host = parts[1]
	} else {
		// 如果没有指定用户，使用当前用户
		// currentUser, err := user.Current()
		// if err != nil {
		// 	return fmt.Errorf("failed to get current user: %v", err)
		// }
		// connArgs.User = currentUser.Username
		connArgs.Host = destination
	}
	return nil
}

// buildSSHCommand 构建 SSH 命令参数
func buildSSHCommand(connArgs ConnArgs) string {
	var args []string

	// 添加端口参数
	if connArgs.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", connArgs.Port))
	}
	// 添加身份文件参数（-i）
	if connArgs.Identity != "" {
		args = append(args, "-i", connArgs.Identity)
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

func init() {
	rootCmd.AddCommand(connectCmd)

	// 用户参数 - 如果同时在 destination 和 -u 中指定，以 -u 为准
	connectCmd.Flags().StringVarP(&connArgs.User, "user", "u", "", "username for SSH connection")

	// 端口参数
	connectCmd.Flags().IntVarP(&connArgs.Port, "port", "p", 22, "SSH port")

	// 身份文件参数（标准 SSH 的 -i 参数）
	connectCmd.Flags().StringVarP(&connArgs.Identity, "identity", "i", "", "identity file (private key) for public key authentication")
}
