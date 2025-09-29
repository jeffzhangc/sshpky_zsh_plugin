// cmd/connect.go
package cmd

import (
	"fmt"
	"os"
	"sshpky/pkg/sshmanager"
	"strings"

	"github.com/spf13/cobra"
)

var (
	user    string
	port    int
	keyName string
)

var connectCmd = &cobra.Command{
	Use:   "connect [host]",
	Short: "Connect to a host using managed SSH keys",
	Long: `Connect to a remote host using one of the managed SSH keys.
This command will automatically select and use the appropriate key.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]

		if err := connectToHost(host, user, port, keyName); err != nil {
			fmt.Printf("Error connecting to host: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVarP(&user, "user", "u", "root", "username for SSH connection")
	connectCmd.Flags().IntVarP(&port, "port", "p", 22, "SSH port")
	connectCmd.Flags().StringVarP(&keyName, "key", "k", "default", "key name to use for connection")
}

func connectToHost(host, user string, port int, keyName string) error {
	// 这里实现SSH连接的逻辑
	fmt.Printf("Connecting to %s@%s:%d using key '%s'\n", user, host, port, keyName)
	// fmt.Printf("Using key from: %s/%s/%s\n", configDir, host, keyName)
	// fmt.Println("SSH connection established successfully!")
	host = "10.1.106.93"
	sshArgs := []string{"-l", "root", "-p", "22", "-tt"}
	pass, _ := sshmanager.GetPassword(host)
	otpSecret, _ := sshmanager.GetMFASecret(host)
	otpCode := ""
	var err error

	if otpSecret != "" {
		otpCode, _ = sshmanager.GenerateOTP(otpSecret)
	}

	// 先尝试用已有密码 + OTP 自动登录
	err = sshmanager.RunSSHWithAutoInput(sshmanager.SSHOptions{
		Host:     host,
		SSHArgs:  sshArgs,
		Password: pass,
		OTPCode:  otpCode,
		Debug:    false,
	})

	// 如果成功，直接返回
	if err == nil {
		fmt.Println("✅ 自动登录成功")
		return nil
	}

	// 如果失败，开始交互输入密码 / OTP

	// 重新获取密码（用户手动输入）
	if pass == "" || strings.Contains(err.Error(), "Permission denied") {
		fmt.Print("请输入密码: ")
		fmt.Scanln(&pass)
	}

	// 重新获取 MFA Secret（用户手动输入）
	if otpSecret == "" || strings.Contains(err.Error(), "Verification code") {
		fmt.Print("请输入 MFA Secret（Google Authenticator 中的 Key）: ")
		fmt.Scanln(&otpSecret)
		// 保存以便下次自动生成
		sshmanager.SaveMFASecret(host, otpSecret)
	}

	// 用新密码 + OTP 再次尝试登录
	otpCode, _ = sshmanager.GenerateOTP(otpSecret)

	err = sshmanager.RunSSHWithAutoInput(sshmanager.SSHOptions{
		Host:     host,
		SSHArgs:  sshArgs,
		Password: pass,
		OTPCode:  otpCode,
		Debug:    false,
	})

	// 登录成功则保存密码
	if err == nil {
		fmt.Println("✅ 登录成功，密码和 OTP 已保存")
		sshmanager.SavePassword(host, pass)
	} else {
		fmt.Fprintf(os.Stderr, "❌ 登录失败：%v\n", err)
		os.Exit(1)
	}

	// 在实际实现中，这里会启动一个SSH会话
	return nil
}
