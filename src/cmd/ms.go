package cmd

import (
	"fmt"
	"os"
	"sshpky/pkg/config"
	"sshpky/pkg/utils"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	searchKeyword string
	showDetail    bool
	groupName     string
)

/**
 *
 * manager ssh 管理、显示 ssh 配置内容
 * 1. list 显示 当前 use group 下所有的 ssh 配置信息
 * 			-s 搜索 模糊匹配
 * 2. list xxx  显示 xxx 组下 所有的 ssh 配置信息
 * 			-s 搜索
 * 3. delete abc 删除 abc 的 ssh 配置信息
 * 4. get abc 显示 abc 的 ssh 配置信息
 *
 */
var msCmd = &cobra.Command{
	Use:   "ms",
	Short: "Manage SSH config items",
	Long: `Manage SSH key groups and configurations.
This command allows you to list, use, and manage different SSH key groups.
You can list all configured SSH key groups, search for specific items, 
delete SSH key configurations, and view detailed information about individual items.`,
	Run: func(cmd *cobra.Command, args []string) {
		msBubble(args)
	},
}

var msListCmd = &cobra.Command{
	Use:               "list [group-name]",
	Short:             "List all ssh client",
	Long:              `List all configured SSH key groups and show the currently active group.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: groupUseValidArgs, // 添加自动补全函数
	Run: func(cmd *cobra.Command, args []string) {
		listSSHConfigs(args, searchKeyword, showDetail)
	},
}

var msDelCmd = &cobra.Command{
	Use:   "delete [host-name]",
	Short: "Delete SSH configuration",
	Long:  `Delete an existing SSH configuration by host name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteSSHConfig(args[0])
	},
}

var msGetCmd = &cobra.Command{
	Use:               "get [host-name]",
	Short:             "Get SSH configuration details",
	Long:              `Get detailed information about a specific SSH configuration.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: msGetCmdValidArgs, // 添加自动补全函数
	Run: func(cmd *cobra.Command, args []string) {
		getSSHConfig(args[0])
	},
}

var msAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new SSH configuration",
	Long:  `Add a new SSH configuration interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		addSSHConfig()
	},
}

var msUpdateCmd = &cobra.Command{
	Use:   "update [host-name]",
	Short: "Update SSH configuration",
	Long:  `Update an existing SSH configuration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		updateSSHConfig(args[0])
	},
}

func init() {
	rootCmd.AddCommand(msCmd)
	msCmd.AddCommand(msListCmd)
	msCmd.AddCommand(msDelCmd)
	msCmd.AddCommand(msGetCmd)
	msCmd.AddCommand(msAddCmd)
	msCmd.AddCommand(msUpdateCmd)

	// 为 list 命令添加标志
	msListCmd.Flags().StringVarP(&searchKeyword, "search", "s", "", "Search keyword for filtering SSH configurations")
	msListCmd.Flags().BoolVarP(&showDetail, "detail", "d", false, "Show detailed configuration information")
	msListCmd.Flags().BoolVar(&noheader, "no-headers", false, "no-headers")
	msGetCmd.Flags().StringVarP(&groupName, "group", "g", "", "group name")
	// 为 group 标志添加补全
	msGetCmd.RegisterFlagCompletionFunc("group", groupFlagValidArgs)
}

func msGetCmdValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return connValidArgs(cmd, args, toComplete)
}

func listSSHConfigs(args []string, searchKeyword string, showDetail bool) {
	manager := config.NewSSHConfigManager("")

	var configs []*config.SshConfigItem
	var err error

	// 确定要显示的组
	groupName := ""
	if len(args) > 0 {
		groupName = args[0]
	} else {
		// 使用当前激活的组
		cfg := config.GetConfig()
		groupName = cfg.Use
	}

	// 如果有搜索关键词，使用搜索功能
	if searchKeyword != "" {
		configs, err = manager.SearchConfigs(searchKeyword)
		if err != nil {
			fmt.Printf("Error searching SSH configurations: %v\n", err)
			os.Exit(1)
		}

		// 如果指定了组，过滤结果
		if groupName != "" {
			var filteredConfigs []*config.SshConfigItem
			for _, config := range configs {
				if config.Group == groupName {
					filteredConfigs = append(filteredConfigs, config)
				}
			}
			configs = filteredConfigs
		}
	} else if groupName != "" {
		// 获取指定组的配置
		configs, err = manager.GetConfigsByGroup(groupName)
		if err != nil {
			fmt.Printf("Error getting SSH configurations for group '%s': %v\n", groupName, err)
			os.Exit(1)
		}
	} else {
		// 获取所有配置
		configs, err = manager.ReadConfig()
		if err != nil {
			fmt.Printf("Error reading SSH configurations: %v\n", err)
			os.Exit(1)
		}
	}

	if len(configs) == 0 {
		if searchKeyword != "" {
			fmt.Printf("No SSH configurations found matching '%s'", searchKeyword)
			if groupName != "" {
				fmt.Printf(" in group '%s'", groupName)
			}
			fmt.Println()
		} else if groupName != "" {
			fmt.Printf("No SSH configurations found in group '%s'\n", groupName)
		} else {
			fmt.Println("No SSH configurations found")
		}
		return
	}

	if !noheader {

		// 显示配置列表
		fmt.Printf("SSH Configurations")
		if groupName != "" {
			fmt.Printf(" in group '%s'", groupName)
		}
		if searchKeyword != "" {
			fmt.Printf(" matching '%s'", searchKeyword)
		}
		fmt.Printf(" (%d found):\n\n", len(configs))
	}

	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	if !noheader {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Host", "HostName", "Port", "User")
	}

	for _, config := range configs {
		// if showDetail {
		// 	printSSHConfigDetail(config, i+1)
		// 	if i < len(configs)-1 {
		// 		fmt.Println("---")
		// 	}
		// } else {
		// 	printSSHConfigSummary(config, i+1)
		// }
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", config.Host, config.HostName, config.Port, config.User)
	}
	fmt.Fprintln(w)
	w.Flush()
}

func deleteSSHConfig(host string) {
	manager := config.NewSSHConfigManager("")

	// 先确认配置是否存在
	config, err := manager.FindConfig(host)
	if err != nil {
		fmt.Printf("Error: SSH configuration for host '%s' not found\n", host)
		os.Exit(1)
	}

	// 显示将要删除的配置信息
	fmt.Printf("The following SSH configuration will be deleted:\n\n")
	printSSHConfigDetail(*config, 0)
	fmt.Println()

	// 确认删除
	confirmStr := utils.GetAnswer("Are you sure you want to delete this SSH configuration?", "", []string{"y", "n"})
	if confirmStr != "y" {
		fmt.Println("Deletion cancelled.")
		return
	}

	// 执行删除
	err = manager.DeleteConfig(host)
	if err != nil {
		fmt.Printf("Error deleting SSH configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SSH configuration for host '%s' deleted successfully\n", host)
}

func getSSHConfig(host string) {
	manager := config.NewSSHConfigManager("")

	config, err := manager.FindConfig(host)
	if err != nil {
		fmt.Printf("Error: SSH configuration for host '%s' not found\n", host)
		os.Exit(1)
	}

	fmt.Printf("SSH Configuration for host '%s':\n\n", host)
	printSSHConfigDetail(*config, 0)
}

func addSSHConfig() {
	manager := config.NewSSHConfigManager("")

	fmt.Println("Add new SSH configuration:")
	fmt.Println("--------------------------")

	// 获取配置信息
	host := utils.GetBaseAnswer("Host alias", "")
	if host == "" {
		fmt.Println("Error: Host alias is required")
		os.Exit(1)
	}

	// 检查是否已存在
	existing, _ := manager.FindConfig(host)
	if existing != nil {
		fmt.Printf("Error: SSH configuration for host '%s' already exists\n", host)
		os.Exit(1)
	}

	hostName := utils.GetBaseAnswer("HostName (IP or domain)", "")
	user := utils.GetBaseAnswer("User", "")
	port := utils.GetBaseAnswer("Port", "22")
	identityFile := utils.GetBaseAnswer("IdentityFile (optional)", "")
	proxyCommand := utils.GetBaseAnswer("ProxyCommand (optional)", "")

	// 获取组信息
	cfg := config.GetConfig()
	group := utils.GetBaseAnswer("Group", cfg.Use)
	desc := utils.GetBaseAnswer("Description", "")

	// 创建配置项
	newConfig := config.SshConfigItem{
		Host:         host,
		HostName:     hostName,
		User:         user,
		Port:         parsePort(port),
		IdentityFile: identityFile,
		ProxyCommand: proxyCommand,
		Group:        group,
		Desc:         desc,
	}

	// 添加配置
	err := manager.AddConfig(newConfig)
	if err != nil {
		fmt.Printf("Error adding SSH configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SSH configuration for host '%s' added successfully\n", host)
}

func updateSSHConfig(host string) {
	manager := config.NewSSHConfigManager("")

	// 获取现有配置
	existing, err := manager.FindConfig(host)
	if err != nil {
		fmt.Printf("Error: SSH configuration for host '%s' not found\n", host)
		os.Exit(1)
	}

	fmt.Printf("Update SSH configuration for host '%s':\n", host)
	fmt.Println("(Leave blank to keep current value)")
	fmt.Println("-----------------------------------")

	// 获取更新信息
	hostName := utils.GetBaseAnswer("HostName", existing.HostName)
	user := utils.GetBaseAnswer("User", existing.User)
	port := utils.GetBaseAnswer("Port", fmt.Sprintf("%d", existing.Port))
	identityFile := utils.GetBaseAnswer("IdentityFile", existing.IdentityFile)
	proxyCommand := utils.GetBaseAnswer("ProxyCommand", existing.ProxyCommand)
	group := utils.GetBaseAnswer("Group", existing.Group)
	desc := utils.GetBaseAnswer("Description", existing.Desc)

	// 更新配置项
	updatedConfig := config.SshConfigItem{
		Host:         host, // 保持原host名不变
		HostName:     hostName,
		User:         user,
		Port:         parsePort(port),
		IdentityFile: identityFile,
		ProxyCommand: proxyCommand,
		Group:        group,
		Desc:         desc,
	}

	// 更新配置
	err = manager.UpdateConfig(host, updatedConfig)
	if err != nil {
		fmt.Printf("Error updating SSH configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SSH configuration for host '%s' updated successfully\n", host)
}

func printSSHConfigSummary(config config.SshConfigItem, index int) {
	indexStr := ""
	if index > 0 {
		indexStr = fmt.Sprintf("%d. ", index)
	}

	fmt.Printf("%sHost: %s\n", indexStr, config.Host)
	if config.HostName != "" {
		fmt.Printf("   HostName: %s\n", config.HostName)
	}
	if config.User != "" {
		fmt.Printf("   User: %s\n", config.User)
	}
	if config.Port != 0 && config.Port != 22 {
		fmt.Printf("   Port: %d\n", config.Port)
	}
	if config.Group != "" {
		fmt.Printf("   Group: %s\n", config.Group)
	}
	if config.Desc != "" {
		fmt.Printf("   Desc: %s\n", config.Desc)
	}
	fmt.Println()
}

func printSSHConfigDetail(config config.SshConfigItem, index int) {
	indexStr := ""
	if index > 0 {
		indexStr = fmt.Sprintf("%d. ", index)
	}

	fmt.Printf("%sHost: %s\n", indexStr, config.Host)
	fmt.Printf("  HostName: %s\n", config.HostName)
	fmt.Printf("  User: %s\n", config.User)
	fmt.Printf("  Port: %d\n", config.Port)
	if config.IdentityFile != "" {
		fmt.Printf("  IdentityFile: %s\n", config.IdentityFile)
	}
	if config.ProxyCommand != "" {
		fmt.Printf("  ProxyCommand: %s\n", config.ProxyCommand)
	}
	if config.Group != "" {
		fmt.Printf("  Group: %s\n", config.Group)
	}
	if config.Desc != "" {
		fmt.Printf("  Description: %s\n", config.Desc)
	}
	if config.EditTime != "" {
		fmt.Printf("  Last Edit: %s\n", config.EditTime)
	}
}

func parsePort(portStr string) int {
	if portStr == "" {
		return 22
	}

	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil || port <= 0 || port > 65535 {
		fmt.Printf("Warning: Invalid port '%s', using default port 22\n", portStr)
		return 22
	}
	return port
}
