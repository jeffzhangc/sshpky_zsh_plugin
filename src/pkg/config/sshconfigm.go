package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SSHConfigManager 管理 SSH 配置文件
type SSHConfigManager struct {
	ConfigPath string
}

var sshConfigM *SSHConfigManager
var once sync.Once
var commonConf []string = []string{}
var sshConfigs []*SshConfigItem = []*SshConfigItem{}

// NewSSHConfigManager 创建新的 SSH 配置管理器
func NewSSHConfigManager(configPath string) *SSHConfigManager {
	once.Do(func() {
		homeDir, _ := os.UserHomeDir()
		if configPath == "" {
			configPath = filepath.Join(homeDir, ".ssh", "config")
		}
		sshConfigM = &SSHConfigManager{
			ConfigPath: configPath,
		}
		sshConfigs, _ = sshConfigM.readConfig()
	})
	return sshConfigM
}

func (m *SSHConfigManager) ReadConfig() ([]*SshConfigItem, error) {
	return sshConfigs, nil
}

// ReadConfig 读取整个 SSH 配置文件
func (m *SSHConfigManager) readConfig() ([]*SshConfigItem, error) {
	var configs []*SshConfigItem

	// 检查文件是否存在
	if _, err := os.Stat(m.ConfigPath); os.IsNotExist(err) {
		return configs, nil // 文件不存在返回空切片
	}

	file, err := os.Open(m.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentConfig *SshConfigItem
	var inHostBlock bool
	var currentComments []string // 存储当前配置块前的注释

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " ")

		// 处理注释行
		if strings.HasPrefix(line, "#") {
			// 提取注释内容（去掉 # 和空格）
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			currentComments = append(currentComments, comment)
			continue
		}

		// 跳过空行
		if line == "" {
			// 空行可能表示配置块结束，但我们等到遇到下一个 Host 时才处理
			continue
		}

		// 检查是否是 Host 块开始
		if strings.HasPrefix(line, "Host ") {
			// 保存前一个配置项
			// if currentConfig != nil {
			// 	// 处理之前收集的注释信息
			// 	m.parseComments(currentComments, currentConfig)
			// 	currentConfig = nil
			// }

			// 开始新的配置项
			hosts := strings.Fields(line)[1:]
			if len(hosts) > 0 {
				currentConfig = &SshConfigItem{
					Host:        hosts[0],
					Port:        22, // 默认端口
					EditTime:    time.Now().Format("2006-01-02 15:04:05"),
					OtherParams: []string{},
				}
				inHostBlock = true
				// 处理当前收集的注释
				m.parseComments(currentComments, currentConfig)
				configs = append(configs, currentConfig)

				currentComments = nil // 重置注释
			}
			// continue
		}

		// 如果在 Host 块中，解析配置项
		if inHostBlock && currentConfig != nil && strings.HasPrefix(line, " ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				key := strings.ToLower(fields[0])
				value := strings.Join(fields[1:], " ")

				switch key {
				case "hostname":
					currentConfig.HostName = value
				case "port":
					if port, err := strconv.Atoi(value); err == nil {
						currentConfig.Port = port
					}
				case "user":
					currentConfig.User = value
				case "identityfile":
					currentConfig.IdentityFile = value
				case "proxycommand":
					currentConfig.ProxyCommand = value
				default:
					currentConfig.OtherParams = append(currentConfig.OtherParams, line)
				}
			}
			continue
		}

		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "Host ") {
			commonConf = append(commonConf, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	return configs, nil
}

// parseComments 从注释中解析 Group、Desc 和 EditTime 信息
func (m *SSHConfigManager) parseComments(comments []string, config *SshConfigItem) {
	for _, comment := range comments {
		// 解析分组信息
		if strings.HasPrefix(comment, "Group:") {
			config.Group = strings.TrimSpace(strings.TrimPrefix(comment, "Group:"))
		} else if strings.HasPrefix(comment, "Desc:") || strings.HasPrefix(comment,
			"Description:") {
			// 解析描述信息
			config.Desc = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(comment, "Desc:"), "Description:"))
		} else if strings.HasPrefix(comment, "EditTime:") || strings.HasPrefix(comment, "Last Edit:") {
			// 解析编辑时间
			config.EditTime = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(comment, "EditTime:"), "Last Edit:"))
		} else if strings.HasPrefix(comment, "Password:") {
			config.Password = strings.TrimSpace(strings.TrimPrefix(comment, "Password:"))
		} else if strings.HasPrefix(comment, "MFASecret") {
			config.MFASecret = strings.TrimSpace(strings.TrimPrefix(comment, "MFASecret:"))
		} else if config.Desc == "" && !strings.Contains(comment, ":") {
			// 如果注释不是特定格式，且没有描述，将其作为描述
			config.Desc = comment
		}
	}
	if config.Group == "" {
		config.Group = DEFAULT_USENAME
	}
}

// AddConfig 添加新的 SSH 配置项
func (m *SSHConfigManager) AddConfig(item SshConfigItem) error {
	// 确保 SSH 目录存在
	sshDir := filepath.Dir(m.ConfigPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("创建 SSH 目录失败: %v", err)
	}

	// 检查是否已存在相同 Host 的配置
	existingConfigs, err := m.ReadConfig()
	if err != nil {
		return err
	}

	for _, config := range existingConfigs {
		if config.Host == item.Host {
			return fmt.Errorf("Host '%s' 已存在", item.Host)
		}
	}

	// 打开文件追加模式
	file, err := os.OpenFile(m.ConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入配置项
	if _, err := writer.WriteString(m.formatConfigItem(item)); err != nil {
		return fmt.Errorf("写入配置失败: %v", err)
	}
	sshConfigs = append(sshConfigs, &item)

	return writer.Flush()
}

// UpdateConfig 更新现有的 SSH 配置项
func (m *SSHConfigManager) UpdateConfig(host string, newItem SshConfigItem) error {
	configs, err := m.ReadConfig()
	if err != nil {
		return err
	}

	found := false
	for i, config := range configs {
		if config.Host == host {
			// 保持原有的 Host 值，除非明确指定要修改
			if newItem.Host == "" {
				newItem.Host = host
			}
			newItem.EditTime = time.Now().Format("2006-01-02 15:04:05")
			configs[i] = &newItem
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到 Host '%s' 的配置", host)
	}

	err = m.writeAllConfigs(configs)
	sshConfigs = configs
	return err
}

// DeleteConfig 删除指定的 SSH 配置项
func (m *SSHConfigManager) DeleteConfig(host string) error {
	configs, err := m.ReadConfig()
	if err != nil {
		return err
	}

	var newConfigs []*SshConfigItem
	found := false

	for _, config := range configs {
		if config.Host == host {
			found = true
			continue
		}
		newConfigs = append(newConfigs, config)
	}

	if !found {
		return fmt.Errorf("未找到 Host '%s' 的配置", host)
	}

	return m.writeAllConfigs(newConfigs)
}

// FindConfig 查找指定的 SSH 配置项
func (m *SSHConfigManager) FindConfig(host string) (*SshConfigItem, error) {
	configs, err := m.ReadConfig()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.Host == host {
			return config, nil
		}
	}

	return nil, fmt.Errorf("未找到 Host '%s' 的配置", host)
}

// GetConfigsByGroup 根据分组获取配置项
func (m *SSHConfigManager) GetConfigsByGroup(group string) ([]*SshConfigItem, error) {
	configs, err := m.ReadConfig()
	if err != nil {
		return nil, err
	}

	var result []*SshConfigItem
	for _, config := range configs {
		if config.Group == group {
			result = append(result, config)
		}
	}

	return result, nil
}

// SearchConfigs 搜索配置项（根据 Host, HostName, Desc 字段）
func (m *SSHConfigManager) SearchConfigs(keyword string) ([]*SshConfigItem, error) {
	configs, err := m.ReadConfig()
	if err != nil {
		return nil, err
	}

	var result []*SshConfigItem
	pattern := strings.ToLower(keyword)

	for _, config := range configs {
		if strings.Contains(strings.ToLower(config.Host), pattern) ||
			strings.Contains(strings.ToLower(config.HostName), pattern) ||
			// strings.Contains(strings.ToLower(config.Desc), pattern)
			strings.Contains(strings.ToLower(config.Group), pattern) {
			result = append(result, config)
		}
	}

	return result, nil
}

// writeAllConfigs 将所有配置项写入文件
func (m *SSHConfigManager) writeAllConfigs(configs []*SshConfigItem) error {
	// 确保 SSH 目录存在
	sshDir := filepath.Dir(m.ConfigPath)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return fmt.Errorf("创建 SSH 目录失败: %v", err)
	}

	file, err := os.Create(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// 写入文件头注释
	header := "# SSH Config File\n# Generated by SSH Config Manager\n# Edit Time: " + time.Now().Format("2006-01-02 15:04:05") + "\n\n"
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("写入文件头失败: %v", err)
	}

	for _, comm := range commonConf {
		writer.WriteString(comm + "\n")
	}

	// 写入所有配置项
	for _, config := range configs {
		if _, err := writer.WriteString(m.formatConfigItem(*config)); err != nil {
			return fmt.Errorf("写入配置失败: %v", err)
		}

	}
	sshConfigs = configs
	return writer.Flush()
}

// formatConfigItem 将配置项格式化为 SSH config 文件格式
func (m *SSHConfigManager) formatConfigItem(item SshConfigItem) string {
	var builder strings.Builder

	// 添加分组注释（如果有）
	if item.Group != "" {
		builder.WriteString(fmt.Sprintf("\n# Group: %s\n", item.Group))
	}

	// 添加描述注释（如果有）
	if item.Desc != "" {
		builder.WriteString(fmt.Sprintf("# Desc: %s\n", item.Desc))
	}

	// 添加编辑时间注释
	if item.EditTime != "" {
		builder.WriteString(fmt.Sprintf("# EditTime: %s\n", item.EditTime))
	}

	// 密码
	if item.Password != "" {
		builder.WriteString(fmt.Sprintf("# Password: %s\n", item.Password))
	}

	// mfa
	if item.MFASecret != "" {
		builder.WriteString(fmt.Sprintf("# MFASecret: %s\n", item.MFASecret))
	}

	// 开始 Host 块
	builder.WriteString(fmt.Sprintf("Host %s\n", item.Host))

	// 添加各项配置
	if item.HostName != "" {
		builder.WriteString(fmt.Sprintf("    HostName %s\n", item.HostName))
	}

	if item.Port != 0 && item.Port != 22 {
		builder.WriteString(fmt.Sprintf("    Port %d\n", item.Port))
	}

	if item.User != "" {
		builder.WriteString(fmt.Sprintf("    User %s\n", item.User))
	}

	if item.IdentityFile != "" {
		builder.WriteString(fmt.Sprintf("    IdentityFile %s\n", item.IdentityFile))
	}

	if item.ProxyCommand != "" {
		builder.WriteString(fmt.Sprintf("    ProxyCommand %s\n", item.ProxyCommand))
	}

	if len(item.OtherParams) > 0 {
		for _, os := range item.OtherParams {
			builder.WriteString(fmt.Sprintf("%s\n", os))
		}
	}

	builder.WriteString("\n")
	return builder.String()
}

// BackupConfig 备份当前的 SSH 配置文件
func (m *SSHConfigManager) BackupConfig() error {
	backupPath := m.ConfigPath + ".backup_" + time.Now().Format("20060102_150405")

	data, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("创建备份文件失败: %v", err)
	}

	return nil
}

// ValidateConfig 验证 SSH 配置文件的语法
func (m *SSHConfigManager) ValidateConfig() error {
	_, err := m.ReadConfig()
	return err
}

// 对外方法，保存 configItem
func SaveConfigFromConn(connConf SshConfigItem) {
	ssm := NewSSHConfigManager("")
	var group *SshpkyGroupConfig = connConf.getGroup()
	if connConf.Group == "" {
		connConf.Group = config.Use
	}

	if connConf.HostName == "" {
		connConf.HostName = connConf.Host
	}

	if group == nil {
		// group 不存在，暂不自动保存
		fmt.Printf("group %s is not exist,do not auto save group\n", connConf.Group)
		return
	}

	existConfig, _ := ssm.FindConfig(connConf.Host)

	// 存储密码
	keym := group.getKeyManager()

	if existConfig != nil && (existConfig.Password != connConf.Password || existConfig.MFASecret != connConf.MFASecret) {
		keym.SavePwd(&connConf)
		connConf.EditTime = time.Now().Format("2006-01-02 15:04:05")
		ssm.UpdateConfig(connConf.Host, connConf)
		fmt.Printf("update %s success\n", connConf.Host)
	} else if existConfig == nil {
		connConf.EditTime = time.Now().Format("2006-01-02 15:04:05")
		keym.SavePwd(&connConf)
		ssm.AddConfig(connConf)
		fmt.Printf("add %s success\n", connConf.Host)
	}
}
