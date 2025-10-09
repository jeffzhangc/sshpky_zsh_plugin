package sshrunner

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sshpky/pkg/config"
	"sshpky/pkg/km"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

type SSHOptions struct {
	Host     string
	SSHArgs  []string
	Password string
	OTPCode  string
	Debug    bool
}

func RunSSH(sshCmd string, conn config.SshConfigItem, args []string) error {
	shell, err := getShell()

	ms := config.NewSSHConfigManager("")
	cnf, _ := ms.FindConfig(conn.Host)
	if cnf != nil {
		// 使用 配置文件中的 conn 信息
		conn = *cnf
		// fmt.Printf("find %s from config\n", conn.Host)
	}

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

	_, err = autoSSHWithLogin(pt, conn)
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

func autoSSHWithLogin(pt *os.File, connConf config.SshConfigItem) (string, error) {
	errChan := make(chan error)
	msgChan := make(chan string)

	host := connConf.HostName

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
						// password, err = km.GetPassword(username, host)
						password = connConf.GetPassword()

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
					go savePwd(connConf, inputOtpSecret, inputPassword)
					msgChan <- data + "\n"
					return
				}

				if strings.Contains(line, "OTP Code") {
					os.Stdout.WriteString(line + ",origianl secret")
					var optSecret string
					if otpTryTime == 0 {
						// optPwd, _ = km.GetMFASecret(username, host)
						optSecret = connConf.GetMafSecret()
					}
					if optSecret == "" {
						// reader := bufio.NewReader(os.Stdin)
						// // fmt.Println("请输入内容：")
						// byteoptSecret, _ := reader.ReadBytes('\n') // 直接读取到换行符
						byteoptSecret, err := term.ReadPassword(int(os.Stdin.Fd()))
						if err != nil {
							errChan <- fmt.Errorf("failed to read password: %v", err)
							return
						}
						optSecret = string(byteoptSecret)
						inputOtpSecret = optSecret
					}
					optPwd, _ := km.GenerateOTP(optSecret)
					inputOtpSecret = optSecret
					data = "" // 清空已处理的数据
					_, err = pt.Write([]byte(optPwd + "\n"))
					if err != nil {
						errChan <- fmt.Errorf("failed to enter optCode: %v", err)
						return
					}
					continue
				}

				if strings.Contains(line, "Host key verification failed") {
					// ssh key error
					os.Stdout.WriteString(data)
					os.Stdout.WriteString("auto remove known_hosts? y/n default:y")
					reader := bufio.NewReader(os.Stdin)
					// fmt.Println("请输入内容：")
					confirmStr, _ := reader.ReadBytes('\n') // 直接读取到换行符
					fmt.Println("write str:", string(confirmStr))
					if confirmStr != nil && string(confirmStr) == "N" {
						errChan <- fmt.Errorf("do not modify knonw_hosts password: %v", err)
						return
					} else {
						// os.Stdout.WriteString("auto modified knonw_hosts,retry to sshpky")
						modifyFixKnowHost(data)
						errChan <- fmt.Errorf("modified knonw_hosts,retry ssh: %v", host)
					}
					data = ""
					return
				}
			}
		}
	}()

	timer := time.NewTimer(time.Second * 10)
	defer timer.Stop()

	select {
	case newBuffered := <-msgChan:
		os.Stdout.WriteString(newBuffered)
		time.Sleep(time.Millisecond * 500)

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

		go func() { _, _ = io.Copy(pt, os.Stdin) }()
		_, _ = io.Copy(os.Stdout, pt)
		// os.Stdout.WriteString("x11" + newBuffered + "xxxx")
		return "", nil
	case err := <-errChan:
		return "", err
	case <-timer.C:
		return "", fmt.Errorf("timed out waiting for prompt")
	}
}
func savePwd(connConf config.SshConfigItem, otpSecret, inputPassword string) {
	// username := connConf.User
	// host := connConf.HostName

	// if inputPassword != "" {
	// 	km.SavePassword(username, host, inputPassword)
	// }

	// if otpSecret != "" {
	// 	km.SaveMFASecret(username, host, otpSecret)
	// }
	if otpSecret != "" {
		connConf.MFASecret = otpSecret
	}
	if inputPassword != "" {
		connConf.Password = inputPassword
	}

	if connConf.Password != "" || connConf.MFASecret != "" {
		config.SaveConfigFromConn(connConf)
	}
}

func modifyFixKnowHost(data string) {
	// 从错误信息中提取主机地址和端口
	host, port, err := extractHostAndPort(data)
	if err != nil {
		fmt.Printf("提取主机信息失败: %v\n", err)
		return
	}

	// 获取 known_hosts 文件路径
	knownHostsPath, err := getKnownHostsPath(data)
	if err != nil {
		fmt.Printf("获取 known_hosts 文件路径失败: %v\n", err)
		return
	}

	fmt.Printf("找到 [%s]:%s 在 %s 文件中，并删除\n", host, port, knownHostsPath)

	// 删除指定主机的记录
	err = removeHostFromKnownHosts(knownHostsPath, host, port)
	if err != nil {
		fmt.Printf("删除记录失败: %v\n", err)
		return
	}

	fmt.Println("成功删除已知主机记录")
}

// 从错误信息中提取主机和端口
func extractHostAndPort(data string) (string, string, error) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Host key for [") && strings.Contains(line, "has changed") {
			// 提取类似 "[10.1.102.34]:5522" 的部分
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start != -1 && end != -1 {
				host := line[start+1 : end]
				// 提取端口
				portStart := strings.Index(line, "]:")
				if portStart != -1 {
					portEnd := strings.Index(line[portStart:], " ")
					if portEnd == -1 {
						portEnd = len(line)
					} else {
						portEnd += portStart
					}
					port := line[portStart+2 : portEnd]
					return host, port, nil
				}
			}
		}
	}
	return "", "", fmt.Errorf("无法从错误信息中提取主机和端口")
}

// 获取 known_hosts 文件路径
func getKnownHostsPath(data string) (string, error) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Offending") && strings.Contains(line, "known_hosts") {
			// 提取文件路径
			start := strings.Index(line, "in ")
			if start != -1 {
				pathPart := line[start+3:]
				end := strings.Index(pathPart, ":")
				if end != -1 {
					return pathPart[:end], nil
				}
				return pathPart, nil
			}
		}
		if strings.Contains(line, "Add correct host key in") {
			// 从另一行提取文件路径
			start := strings.Index(line, "in ")
			if start != -1 {
				pathPart := line[start+3:]
				end := strings.Index(pathPart, " to")
				if end != -1 {
					return pathPart[:end], nil
				}
				return pathPart, nil
			}
		}
	}

	// 如果无法从错误信息中提取，使用默认路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %v", err)
	}
	return homeDir + "/.ssh/known_hosts", nil
}

// 从 known_hosts 文件中删除指定主机的记录
func removeHostFromKnownHosts(filePath, host, port string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 读取文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	targetPattern1 := fmt.Sprintf("[%s]:%s", host, port)
	targetPattern2 := fmt.Sprintf("%s:%s", host, port)

	for scanner.Scan() {
		line := scanner.Text()
		// 跳过以目标主机开头的行（两种格式都检查）
		if strings.HasPrefix(line, targetPattern1) || strings.HasPrefix(line, targetPattern2) {
			fmt.Printf("删除记录: %s\n", line)
			continue
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 写回文件
	return writeLinesToFile(filePath, lines)
}

// 将行写回文件
func writeLinesToFile(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}
	return writer.Flush()
}
