package sshrunner

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sshpky/pkg/km"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/riywo/loginshell"
	"golang.org/x/term"
)

type SSHOptions struct {
	Host     string
	SSHArgs  []string
	Password string
	OTPCode  string
	Debug    bool
}

func RunSSH(sshCmd string, username string, host string, port int) error {
	shell, err := loginshell.Shell()
	if err != nil {
		shell = "/bin/bash"
	}
	c := exec.Command(shell, "-f")

	c.Env = append(os.Environ(), "HISTFILE=/dev/null", "HISTSIZE=0", "HISTFILESIZE=0")
	pt, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = pt.Close() }()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, pt); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGHUP

	errChan := make(chan error, 10)

	go func() {
		// if _, err := pt.Write([]byte("unset HISTFILE; export HISTSIZE=0 \n " + sshCmd + ";exit\n")); err != nil {
		if _, err := pt.Write([]byte(sshCmd + ";exit\n")); err != nil {
			errChan <- err
		}
	}()

	_, err = autoSSHWithLogin(pt, username, host)
	errChan <- err
	select {
	case er := <-errChan:
		return er
	}
	// // fmt.Println(buf, err)
	// if err != nil {
	// 	fmt.Println("runssh error", buf, err)
	// 	return err
	// }
	// return nil
}

func autoSSHWithLogin(pt *os.File, username, host string) (string, error) {
	errChan := make(chan error)
	msgChan := make(chan string)

	var (
		inputPassword    string
		inputOtpSecret   string
		autoTryLoginTime int
		otpTryTime       int
	)

	go func() {

		var data string
		// var err error
		// reader := bufio.NewReader(pt)
		for {
			buf := make([]byte, 4096)
			// n, err := pt.Read(buf)
			n, err := pt.Read(buf)
			// data, err = reader.ReadString('\n')
			if err != nil {
				errChan <- err
				break
			}
			if n == 0 {
				continue
			}
			// 检查是否包含回车符或行结束
			data += string(buf[:n])

			for _, line := range strings.Split(data, "\n") {
				line = strings.Trim(line, " ")
				line = strings.Trim(line, "\t")
				// 处理主机认证确认
				if strings.Contains(line, "The authenticity of host") {
					// 自动确认主机指纹
					_, err := pt.Write([]byte("yes\n"))
					if err != nil {
						errChan <- fmt.Errorf("failed to confirm host: %v", err)
						return
					}
					data = "" // 清空已处理的数据
					continue
				}

				// 处理密码提示
				if strings.Contains(strings.ToLower(line), "password") ||
					strings.Contains(line, "Enter passphrase") ||
					strings.Contains(line, "Password:") ||
					strings.Contains(line, "password:") {
					os.Stdout.WriteString(line)

					var password string
					var err error
					if autoTryLoginTime == 0 {
						password, err = km.GetPassword(username, host)
						if err != nil {
							errChan <- err
							return
						}
						if password != "" && len(password) > 0 {
							autoTryLoginTime += 1
						}
					}

					if password == "" || len(password) == 0 || autoTryLoginTime > 1 {
						// fmt.Println("output:", data, "wait for input password")
						// 输入密码
						// 从终端读取，禁用回显
						bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
						if err != nil {
							errChan <- fmt.Errorf("failed to read password: %v", err)
							return
						}
						password = string(bytePassword)
						inputPassword = password
					}
					// fmt.Println() // 在密码输入后换行
					os.Stdout.WriteString("\n")
					_, err = pt.Write([]byte(password + "\n"))
					if err != nil {
						errChan <- fmt.Errorf("failed to enter password: %v", err)
						return
					}
					data = "" // 清空已处理的数据
					continue
				}

				// 检查认证失败
				if strings.Contains(line, "Permission denied") ||
					strings.Contains(line, "Authentication failed") ||
					strings.Contains(line, "Access denied") {
					errChan <- fmt.Errorf("authentication failure")
					return
				}

				// 检查认证成功 - 出现命令提示符或成功连接
				if strings.Contains(line, "$") ||
					strings.Contains(line, "#") ||
					strings.Contains(line, ">") ||
					strings.Contains(line, "Last login") ||
					strings.Contains(line, "欢迎") ||
					strings.Contains(line, "Welcome") {
					os.Stdout.WriteString("login success\r\n")
					go savePwd(username, host, inputOtpSecret, inputPassword)
					msgChan <- line + "\n"
					return
				}

				if strings.Contains(line, "OTP Code") {
					os.Stdout.WriteString(line + ",origianl secret")
					var optPwd string
					if otpTryTime == 0 {
						optPwd, _ = km.GetMFASecret(username, host)
					}
					if optPwd == "" {
						// reader := bufio.NewReader(os.Stdin)
						// // fmt.Println("请输入内容：")
						// byteoptSecret, _ := reader.ReadBytes('\n') // 直接读取到换行符
						byteoptSecret, err := term.ReadPassword(int(os.Stdin.Fd()))
						if err != nil {
							errChan <- fmt.Errorf("failed to read password: %v", err)
							return
						}
						optSecret := string(byteoptSecret)
						optPwd, _ = km.GenerateOTP(optSecret)
						inputOtpSecret = optSecret
					}
					data = "" // 清空已处理的数据
					_, err = pt.Write([]byte(optPwd + "\n"))
					if err != nil {
						errChan <- fmt.Errorf("failed to enter optCode: %v", err)
						return
					}

					continue
				}
			}
		}
	}()

	timer := time.NewTimer(time.Second * 10)
	defer timer.Stop()

	select {
	case newBuffered := <-msgChan:
		os.Stdout.WriteString(newBuffered)

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

		go func() { _, _ = io.Copy(pt, os.Stdin) }()
		_, _ = io.Copy(os.Stdout, pt)
		return "", nil
	case err := <-errChan:
		return "", err
	case <-timer.C:
		return "", fmt.Errorf("timed out waiting for prompt")
	}
}

func savePwd(username, host, otpSecret, inputPassword string) {
	if inputPassword != "" {
		km.SavePassword(username, host, inputPassword)
	}

	if otpSecret != "" {
		km.SaveMFASecret(username, host, otpSecret)
	}
}
