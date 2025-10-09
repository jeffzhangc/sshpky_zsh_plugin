package cmd

import (
	"fmt"
	"os"
	"sshpky/pkg/config"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 定义模型状态
type state int

const (
	stateTable state = iota
	stateDetail
	stateDeleteConfirm
	stateSearch
	stateAddForm
	stateUpdateForm
	stateQuitWithConn
)

// 表单字段索引
const (
	hostField = iota
	hostNameField
	userField
	portField
	identityFileField
	proxyCommandField
	groupField
	descField
)

// 主模型
type msModel struct {
	state           state
	table           table.Model
	searchInput     textinput.Model
	formInputs      []textinput.Model
	configs         []*config.SshConfigItem
	filteredConfigs []*config.SshConfigItem
	currentConfig   *config.SshConfigItem
	selectedIndex   int
	searchKeyword   string
	groupName       string
	width           int
	height          int
	err             error
	formTitle       string
	focusedField    int
	isNewConfig     bool
	showPassword    bool // 是否显示密码
}

// 表格样式
var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B9D")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	formStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	fieldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Width(40)

	passwordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D26A")).
			Bold(true)

	hiddenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

// 初始化模型
func initialModel(groupName string) msModel {
	// 加载配置
	manager := config.NewSSHConfigManager("")
	var configs []*config.SshConfigItem
	var err error

	if groupName != "" {
		configs, err = manager.GetConfigsByGroup(groupName)
	} else {
		cfg := config.GetConfig()
		if cfg.Use != "" {
			configs, err = manager.GetConfigsByGroup(cfg.Use)
			groupName = cfg.Use
		} else {
			configs, err = manager.ReadConfig()
		}
	}

	if err != nil {
		return msModel{err: err}
	}
	// 创建表格列
	columns := []table.Column{
		{Title: "Host", Width: 20},
		{Title: "HostName", Width: 25},
		{Title: "User", Width: 15},
		{Title: "Port", Width: 8},
		{Title: "Group", Width: 15},
	}

	// 创建表格行
	rows := make([]table.Row, len(configs))
	for i, config := range configs {
		rows[i] = table.Row{
			config.Host,
			config.HostName,
			config.User,
			fmt.Sprintf("%d", config.Port),
			config.Group,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(20, len(rows))),
	)

	// 设置表格样式
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FF6B9D")).
		Background(lipgloss.Color("235")).
		Bold(false)
	t.SetStyles(s)

	// 搜索输入框
	searchInput := textinput.New()
	searchInput.Placeholder = "Search SSH configurations..."
	searchInput.Focus()

	// 初始化表单输入框
	formInputs := make([]textinput.Model, 8)

	// Host
	formInputs[hostField] = textinput.New()
	formInputs[hostField].Width = 40
	formInputs[hostField].Placeholder = "e.g., myserver"
	formInputs[hostField].Focus()

	// HostName
	formInputs[hostNameField] = textinput.New()
	formInputs[hostNameField].Width = 40
	formInputs[hostNameField].Placeholder = "e.g., 192.168.1.1 or example.com"

	// User
	formInputs[userField] = textinput.New()
	formInputs[userField].Placeholder = "e.g., root"

	// Port
	formInputs[portField] = textinput.New()
	formInputs[portField].Placeholder = "22"
	formInputs[portField].SetValue("22")
	formInputs[portField].Width = 40

	// IdentityFile
	formInputs[identityFileField] = textinput.New()
	formInputs[identityFileField].Width = 40
	formInputs[identityFileField].Placeholder = "e.g., ~/.ssh/id_rsa (optional)"

	// ProxyCommand
	formInputs[proxyCommandField] = textinput.New()
	formInputs[proxyCommandField].Width = 40
	formInputs[proxyCommandField].Placeholder = "e.g., ssh gateway -W %h:%p (optional)"

	// Group
	formInputs[groupField] = textinput.New()
	cfg := config.GetConfig()
	gropus := cfg.GetGroupNames()
	formInputs[groupField].Width = 40
	formInputs[groupField].Placeholder = "e.g., " + strings.Join(gropus, ",")
	if cfg.Use != "" {
		formInputs[groupField].SetValue(cfg.Use)
	}

	// Description
	formInputs[descField] = textinput.New()
	formInputs[descField].Placeholder = "e.g., Production server (optional)"

	return msModel{
		state:           stateTable,
		table:           t,
		searchInput:     searchInput,
		formInputs:      formInputs,
		configs:         configs,
		filteredConfigs: configs,
		groupName:       groupName,
		selectedIndex:   0,
		focusedField:    0,
	}
}

// Bubble Tea 接口实现
func (m msModel) Init() tea.Cmd {
	return nil
}

func (m msModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateTable:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if len(m.filteredConfigs) > 0 {
					selectedIndex := m.table.Cursor()
					if selectedIndex < len(m.filteredConfigs) {
						m.currentConfig = m.filteredConfigs[selectedIndex]
						m.state = stateDetail
					}
				}
			case "d":
				if len(m.filteredConfigs) > 0 {
					selectedIndex := m.table.Cursor()
					if selectedIndex < len(m.filteredConfigs) {
						m.currentConfig = m.filteredConfigs[selectedIndex]
						m.state = stateDeleteConfirm
					}
				}
			case "/":
				m.state = stateSearch
				m.searchInput.SetValue("")
				return m, nil
			case "n":
				m.addNewConfig()
			case "u":
				if len(m.filteredConfigs) > 0 {
					selectedIndex := m.table.Cursor()
					if selectedIndex < len(m.filteredConfigs) {
						m.updateConfig(m.filteredConfigs[selectedIndex])
					}
				}
			case "c":
				if len(m.filteredConfigs) > 0 {
					selectedIndex := m.table.Cursor()
					if selectedIndex < len(m.filteredConfigs) {
						m.currentConfig = m.filteredConfigs[selectedIndex]
						m.state = stateQuitWithConn
					}
					return m, tea.Quit
				}
				// try to connect to current host
				// if m.currentConfig != nil {
				// 	fmt.Println("test....xxx", "try to connect", m.currentConfig.Host)

				// }
			}

		case stateDetail:
			switch msg.String() {
			case "q", "esc", "backspace":
				m.state = stateTable
			case "u":
				if m.currentConfig != nil {
					m.updateConfig(m.currentConfig)
				}
			case "c":
				if m.currentConfig != nil {
					m.state = stateQuitWithConn
					return m, tea.Quit
				}
			case "p":
				if m.currentConfig != nil {
					m.showPassword = !m.showPassword
					// 	// 显示 password
					// 	pwd := m.currentConfig.GetPassword()
					// 	mafPwd := m.currentConfig.GetMafSecret()
					// 	fmt.Println("Password:", pwd)
					// 	fmt.Println("MFASecret:", mafPwd)
				}
			}
		case stateDeleteConfirm:
			switch msg.String() {
			case "y", "Y":
				if m.currentConfig != nil {
					m.deleteConfig(m.currentConfig.Host)
					m.state = stateTable
				}
			case "n", "N", "q", "esc":
				m.state = stateTable
				m.currentConfig = nil
			}
		case stateSearch:
			switch msg.String() {
			case "enter":
				m.searchKeyword = m.searchInput.Value()
				m.filterConfigs()
				m.state = stateTable
			case "esc":
				m.state = stateTable
			}
		case stateAddForm, stateUpdateForm:
			switch msg.String() {
			case "ctrl+s", "enter":
				if m.focusedField == len(m.formInputs)-1 {
					// 最后一个字段，保存配置
					if m.saveConfig() {
						m.state = stateTable
						m.reloadConfigs()
					}
				} else {
					// 移动到下一个字段
					m.nextField()
				}
			case "tab":
				m.nextField()
			case "shift+tab":
				m.prevField()
			case "esc", "ctrl+c":
				m.state = stateTable
				m.resetForm()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(min(20, len(m.filteredConfigs)+2))
	}

	// 更新当前状态的组件
	switch m.state {
	case stateTable:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case stateSearch:
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	case stateAddForm, stateUpdateForm:
		for i := range m.formInputs {
			if i == m.focusedField {
				m.formInputs[i], cmd = m.formInputs[i].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m msModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit", m.err)
	}

	var b strings.Builder

	// 标题
	title := "SSH Configurations"
	if m.groupName != "" {
		title += " - Group: " + m.groupName
	}
	if m.searchKeyword != "" {
		title += " (Search: " + m.searchKeyword + ")"
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")

	switch m.state {
	case stateTable:
		b.WriteString(baseStyle.Render(m.table.View()) + "\n\n")
		b.WriteString(m.helpView())

	case stateDetail:
		if m.currentConfig != nil {
			b.WriteString(m.configDetailView())
		}

	case stateDeleteConfirm:
		if m.currentConfig != nil {
			b.WriteString(m.deleteConfirmView())
		}

	case stateSearch:
		b.WriteString("Search: " + m.searchInput.View() + "\n")
		b.WriteString(helpStyle.Render("Press Enter to search, ESC to cancel"))

	case stateAddForm, stateUpdateForm:
		b.WriteString(m.formView())
	}

	return b.String()
}

// 表单视图
func (m msModel) formView() string {
	var b strings.Builder

	b.WriteString(formStyle.Render(m.formTitle + "\n"))

	fields := []struct {
		label string
		index int
	}{
		{"Host*", hostField},
		{"HostName*", hostNameField},
		{"User*", userField},
		{"Port*", portField},
		{"IdentityFile", identityFileField},
		{"ProxyCommand", proxyCommandField},
		{"Group*", groupField},
		{"Description", descField},
	}

	for _, field := range fields {
		input := m.formInputs[field.index]
		label := field.label

		// 标记必填字段
		if strings.HasSuffix(label, "*") {
			label = fieldStyle.Render(label)
		}

		b.WriteString(fmt.Sprintf("\n%s:", label))
		if field.index == m.focusedField {
			b.WriteString(inputStyle.Render(input.View()) + "\n\n")
		} else {
			b.WriteString(inputStyle.Render(input.View()) + "\n\n")
		}
	}

	// 帮助信息
	helpText := "Tab/Enter: Next field • Shift+Tab: Previous field • Ctrl+S: Save • ESC: Cancel"
	if m.focusedField == len(m.formInputs)-1 {
		helpText += " • Enter: Save"
	}
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

// 帮助信息视图
func (m msModel) helpView() string {
	helpText := "↑↓: Navigate • Enter: View details • d: Delete • n: New • u: Update • /: Search • q: Quit"
	if len(m.filteredConfigs) == 0 {
		helpText += "\nNo configurations found"
	} else {
		helpText += fmt.Sprintf("\n%d configuration(s) found", len(m.filteredConfigs))
	}
	return helpStyle.Render(helpText)
}

// 配置详情视图
func (m msModel) configDetailView() string {
	if m.currentConfig == nil {
		return "No configuration selected"
	}

	config := m.currentConfig
	var b strings.Builder

	b.WriteString(selectedStyle.Render("SSH Configuration Details") + "\n\n")
	b.WriteString(fmt.Sprintf("Host: %s\n", config.Host))
	b.WriteString(fmt.Sprintf("HostName: %s\n", config.HostName))
	b.WriteString(fmt.Sprintf("User: %s\n", config.User))
	b.WriteString(fmt.Sprintf("Port: %d\n", config.Port))

	if config.IdentityFile != "" {
		b.WriteString(fmt.Sprintf("IdentityFile: %s\n", config.IdentityFile))
	}
	if config.ProxyCommand != "" {
		b.WriteString(fmt.Sprintf("ProxyCommand: %s\n", config.ProxyCommand))
	}
	if config.Group != "" {
		b.WriteString(fmt.Sprintf("Group: %s\n", config.Group))
	}
	if config.Desc != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", config.Desc))
	}
	if config.EditTime != "" {
		b.WriteString(fmt.Sprintf("Last Edit: %s\n", config.EditTime))
	}

	// 密码信息显示逻辑
	b.WriteString("\n")
	if m.showPassword {
		// 显示密码信息
		pwd := config.GetPassword()
		mafPwd := config.GetMafSecret()

		b.WriteString(passwordStyle.Render("Authentication Information:") + "\n")
		if pwd != "" {
			b.WriteString(fmt.Sprintf("  Password: %s\n", pwd))
		} else {
			b.WriteString("  Password: [Not set]\n")
		}
		if mafPwd != "" {
			b.WriteString(fmt.Sprintf("  MFASecret: %s\n", mafPwd))
		} else {
			b.WriteString("  MFASecret: [Not set]\n")
		}
		b.WriteString(hiddenStyle.Render("  [Press 'p' to hide]") + "\n")
	} else {
		// 隐藏密码信息
		b.WriteString(hiddenStyle.Render("Authentication Information: [Press 'p' to show]") + "\n")
	}

	if len(config.OtherParams) > 0 {
		b.WriteString("\nOtherParams:\n")
		b.WriteString(strings.Join(config.OtherParams, "\n"))
	}

	// 更新帮助信息，包含密码切换提示
	helpText := "q/ESC: Back to list • u: Update this configuration • c: Connect"
	if m.showPassword {
		helpText += " • p: Hide password"
	} else {
		helpText += " • p: Show password"
	}

	b.WriteString("\n" + helpStyle.Render(helpText))

	// if len(config.OtherParams) > 0 {
	// 	b.WriteString("OtherParams:\n")
	// 	b.WriteString(strings.Join(config.OtherParams, "\n"))
	// }

	// b.WriteString("\n" + helpStyle.Render("q/ESC: Back to list • u: Update this configuration"))

	return b.String()
}

// 删除确认视图
func (m msModel) deleteConfirmView() string {
	if m.currentConfig == nil {
		return "No configuration selected"
	}

	var b strings.Builder
	b.WriteString(selectedStyle.Render("⚠️  Delete Confirmation") + "\n\n")
	b.WriteString(fmt.Sprintf("Are you sure you want to delete SSH configuration for host '%s'?\n\n", m.currentConfig.Host))

	// 显示简化的配置信息
	config := m.currentConfig
	b.WriteString(fmt.Sprintf("Host: %s\n", config.Host))
	b.WriteString(fmt.Sprintf("HostName: %s\n", config.HostName))
	b.WriteString(fmt.Sprintf("User: %s\n", config.User))
	if config.Desc != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", config.Desc))
	}

	b.WriteString("\n" + helpStyle.Render("y: Confirm deletion • n: Cancel"))

	return b.String()
}

// 添加新配置
func (m *msModel) addNewConfig() {
	m.state = stateAddForm
	m.formTitle = "Add New SSH Configuration"
	m.isNewConfig = true
	m.resetForm()
	m.focusedField = 0
	m.formInputs[m.focusedField].Focus()
}

// 更新配置
func (m *msModel) updateConfig(config *config.SshConfigItem) {
	m.state = stateUpdateForm
	m.formTitle = fmt.Sprintf("Update SSH Configuration: %s", config.Host)
	m.isNewConfig = false
	m.currentConfig = config
	m.resetForm()

	// 填充现有数据
	m.formInputs[hostField].SetValue(config.Host)
	m.formInputs[hostNameField].SetValue(config.HostName)
	m.formInputs[userField].SetValue(config.User)
	m.formInputs[portField].SetValue(fmt.Sprintf("%d", config.Port))
	m.formInputs[identityFileField].SetValue(config.IdentityFile)
	m.formInputs[proxyCommandField].SetValue(config.ProxyCommand)
	m.formInputs[groupField].SetValue(config.Group)
	m.formInputs[descField].SetValue(config.Desc)

	m.focusedField = 0
	// time.Sleep(time.Millisecond * 200)
	m.formInputs[m.focusedField].Focus()
}

// 保存配置
func (m *msModel) saveConfig() bool {
	// 验证必填字段
	if m.formInputs[hostField].Value() == "" {
		m.err = fmt.Errorf("Host is required")
		return false
	}
	if m.formInputs[hostNameField].Value() == "" {
		m.err = fmt.Errorf("HostName is required")
		return false
	}
	if m.formInputs[userField].Value() == "" {
		m.err = fmt.Errorf("User is required")
		return false
	}

	// 解析端口
	portStr := m.formInputs[portField].Value()
	if portStr == "" {
		portStr = "22"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		m.err = fmt.Errorf("Invalid port number: %s", portStr)
		return false
	}

	// 创建配置对象
	configItem := config.SshConfigItem{
		Host:         m.formInputs[hostField].Value(),
		HostName:     m.formInputs[hostNameField].Value(),
		User:         m.formInputs[userField].Value(),
		Port:         port,
		IdentityFile: m.formInputs[identityFileField].Value(),
		ProxyCommand: m.formInputs[proxyCommandField].Value(),
		Group:        m.formInputs[groupField].Value(),
		Desc:         m.formInputs[descField].Value(),
	}

	manager := config.NewSSHConfigManager("")

	if m.isNewConfig {
		// 检查是否已存在
		existing, _ := manager.FindConfig(configItem.Host)
		if existing != nil {
			m.err = fmt.Errorf("SSH configuration for host '%s' already exists", configItem.Host)
			return false
		}

		// 添加新配置
		err = manager.AddConfig(configItem)
		if err != nil {
			m.err = fmt.Errorf("Error adding SSH configuration: %v", err)
			return false
		}
	} else {
		// 更新现有配置
		originalHost := m.currentConfig.Host
		err = manager.UpdateConfig(originalHost, configItem)
		if err != nil {
			m.err = fmt.Errorf("Error updating SSH configuration: %v", err)
			return false
		}
	}

	m.err = nil
	m.resetForm()
	return true
}

// 重置表单
func (m *msModel) resetForm() {
	for i := range m.formInputs {
		m.formInputs[i].SetValue("")
		m.formInputs[i].Blur()
	}
	// 重置端口默认值
	m.formInputs[portField].SetValue("22")
	// 重置组默认值
	cfg := config.GetConfig()
	if cfg.Use != "" {
		m.formInputs[groupField].SetValue(cfg.Use)
	}
	m.focusedField = 0
}

// 移动到下一个字段
func (m *msModel) nextField() {
	m.formInputs[m.focusedField].Blur()
	m.focusedField = (m.focusedField + 1) % len(m.formInputs)
	m.formInputs[m.focusedField].Focus()
}

// 移动到上一个字段
func (m *msModel) prevField() {
	m.formInputs[m.focusedField].Blur()
	m.focusedField = (m.focusedField - 1 + len(m.formInputs)) % len(m.formInputs)
	m.formInputs[m.focusedField].Focus()
}

// 过滤配置
func (m *msModel) filterConfigs() {
	if m.searchKeyword == "" {
		m.filteredConfigs = m.configs
	} else {
		var filtered []*config.SshConfigItem
		keyword := strings.ToLower(m.searchKeyword)
		for _, config := range m.configs {
			if strings.Contains(strings.ToLower(config.Host), keyword) ||
				strings.Contains(strings.ToLower(config.HostName), keyword) ||
				strings.Contains(strings.ToLower(config.User), keyword) ||
				strings.Contains(strings.ToLower(config.Group), keyword) ||
				strings.Contains(strings.ToLower(config.Desc), keyword) {
				filtered = append(filtered, config)
			}
		}
		m.filteredConfigs = filtered
	}

	// 更新表格行
	rows := make([]table.Row, len(m.filteredConfigs))
	for i, config := range m.filteredConfigs {
		rows[i] = table.Row{
			config.Host,
			config.HostName,
			config.User,
			fmt.Sprintf("%d", config.Port),
			config.Group,
		}
	}
	m.table.SetRows(rows)
	m.table.SetHeight(min(20, len(rows)+2))
	m.table.GotoTop()
}

// 删除配置
func (m *msModel) deleteConfig(host string) {
	manager := config.NewSSHConfigManager("")
	err := manager.DeleteConfig(host)
	if err != nil {
		m.err = err
		return
	}

	// 重新加载配置
	m.reloadConfigs()
}

// 重新加载配置
func (m *msModel) reloadConfigs() {
	manager := config.NewSSHConfigManager("")
	var err error
	if m.groupName != "" {
		m.configs, err = manager.GetConfigsByGroup(m.groupName)
	} else {
		m.configs, err = manager.ReadConfig()
	}
	if err != nil {
		m.err = err
		return
	}
	m.filterConfigs()
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 更新 msBubble 函数
func msBubble(args []string) {
	groupName := ""
	if len(args) > 0 {
		groupName = args[0]
	}

	model := initialModel(groupName)
	if model.err != nil {
		fmt.Printf("Error loading SSH configurations: %v\n", model.err)
		return
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	runModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	lastM := runModel.(msModel)
	if lastM.state == stateQuitWithConn {
		// fmt.Println("bubble tea quit", lastM.currentConfig)
		if lastM.currentConfig != nil {
			conf := lastM.currentConfig
			fmt.Println("curr config", conf.User, conf.Host, conf.HostName)
			runConn(*conf, []string{})
		}
	}
}
