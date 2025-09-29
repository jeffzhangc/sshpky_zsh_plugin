package sshmanager

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type SSHOptions struct {
	Host     string
	SSHArgs  []string
	Password string
	OTPCode  string
	Debug    bool
}

func RunSSHWithAutoInput(opts SSHOptions) error {
	args := append(opts.SSHArgs, opts.Host)
	cmd := exec.Command("ssh", args...)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// 合并输出
	merged := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(merged)

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println("xxxx", line) // 直接打印用户看见的提示

			// 如果提示输入密码
			if strings.Contains(strings.ToLower(line), "password") {
				if opts.Debug {
					fmt.Println("提示检测：密码")
				}
				stdin.Write([]byte(opts.Password + "\r\n"))
			}

			// 如果提示输入 OTP / MFA 验证码
			if strings.Contains(strings.ToLower(line), "verification code") ||
				strings.Contains(strings.ToLower(line), "otp") {
				if opts.Debug {
					fmt.Println("提示检测：MFA")
				}
				stdin.Write([]byte(opts.OTPCode + "\n"))
			}
		}
	}()

	return cmd.Wait()
}
