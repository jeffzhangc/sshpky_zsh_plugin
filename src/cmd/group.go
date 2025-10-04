package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"sshpky/pkg/config"
	"sshpky/pkg/utils"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var noheader bool

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage SSH key groups",
	Long: `Manage groups for SSH keys configuration.
This command allows you to list, use, and manage different SSH key groups.`,
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long:  `List all configured SSH key groups and show the currently active group.`,
	Run: func(cmd *cobra.Command, args []string) {
		listGroups(noheader)
	},
}

var groupUseCmd = &cobra.Command{
	Use:   "use [group-name]",
	Short: "Set default group",
	Long:  `Set the default group to be used for SSH connections.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 && args[0] == "compline" {

		} else {
			useGroup(args[0])
		}
	},
}

var groupAddCmd = &cobra.Command{
	Use:   "add [group-name]",
	Short: "Add a new group",
	Long:  `Add a new SSH key group configuration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addGroup(args[0])
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete [group-name]",
	Short: "Delete a group",
	Long:  `Delete an existing SSH key group.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteGroup(args[0])
	},
}

func init() {

	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupUseCmd)
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupDeleteCmd)

	groupListCmd.Flags().BoolVar(&noheader, "no-headers", false, "no-headers")
}

func listGroups(noheader bool) {
	cfg := config.GetConfig()

	// fmt.Printf("Current default group: %s\n\n", cfg.Use)
	// fmt.Println("Available groups:")
	// fmt.Println("----------------")

	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	if !noheader {
		fmt.Fprintf(w, "%s\t%s\t%s\n", "Name", "Category", "Description")
	}
	// for _, view := range jj.GetBundle(env).Views {
	// 	//fmt.Printf("views: %+v",view)
	// 	for _, j := range view.Jobs {
	// 		fmt.Fprintf(w, "%s\t%s\n", j.Name, j.URL)
	// 	}
	// }

	for _, group := range cfg.Groups {
		// active := ""
		// if group.Name == cfg.Use {
		// 	active = " (active)"
		// }
		// fmt.Printf("Name: %s%s\n", group.Name, active)
		// fmt.Printf("  Description: %s\n", group.Desc)
		// fmt.Printf("  Auto-save: %v\n", group.AutoSave)
		// fmt.Printf("  Secret: %s\n", maskSecret(group.Secret))
		// fmt.Println()
		fmt.Fprintf(w, "%s\t%s\t%s\n", group.Name, group.Category.String(), group.Desc)
	}
	fmt.Fprintln(w)
	w.Flush()
}

func useGroup(groupName string) {
	cfg := config.GetConfig()

	// 检查group是否存在
	found := false
	for _, group := range cfg.Groups {
		if group.Name == groupName {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Error: Group '%s' not found\n", groupName)
		os.Exit(1)
	}

	cfg.Use = groupName
	config.SetConfig(cfg)
	fmt.Printf("Default group set to: %s\n", groupName)
}

func addGroup(groupName string) {
	cfg := config.GetConfig()

	// 检查group是否已存在
	for _, group := range cfg.Groups {
		if group.Name == groupName {
			fmt.Printf("Error: Group '%s' already exists\n", groupName)
			os.Exit(1)
		}
	}
	category := utils.GetAnswer("Category:", "Store", []string{"StoreKeyChain", "StoreFile"})
	categoryEnum, err := config.ParseSecretCategory(category)
	if err != nil {
		panic(err)
	}

	var secretStr string
	if categoryEnum == config.StoreFile {
		keySize := utils.GetAnswer("KeySize:", "24", []string{"24", "36"})
		s, e := strconv.Atoi(keySize)
		if e != nil {
			panic(e)
		}
		secretB, e := utils.GenerateRandomKey(s)
		if e != nil {
			panic(e)
		}

		secretStr = base64.RawStdEncoding.EncodeToString(secretB)
	}

	desc := utils.GetBaseAnswer("Description:", "")
	if desc == "" {
		desc = fmt.Sprintf("SSH key group %s", groupName)
	}

	// 创建新group
	newGroup := config.SshpkyGroupConfig{
		Name:     groupName,
		Secret:   secretStr, // 初始为空，后续可以设置
		AutoSave: true,
		Desc:     desc,
		Category: categoryEnum,
	}

	cfg.Groups = append(cfg.Groups, newGroup)
	config.SetConfig(cfg)
	fmt.Printf("Group '%s' added successfully\n", groupName)
}

func deleteGroup(groupName string) {
	cfg := config.GetConfig()

	if groupName == cfg.Use {
		fmt.Printf("Error: Cannot delete active group '%s'. Switch to another group first.\n", groupName)
		os.Exit(1)
	}

	// 查找并删除group
	newGroups := []config.SshpkyGroupConfig{}
	found := false

	for _, group := range cfg.Groups {
		if group.Name == groupName {
			found = true
			continue
		}
		newGroups = append(newGroups, group)
	}

	if !found {
		fmt.Printf("Error: Group '%s' not found\n", groupName)
		os.Exit(1)
	}

	cfg.Groups = newGroups
	config.SetConfig(cfg)
	fmt.Printf("Group '%s' deleted successfully\n", groupName)
}

func maskSecret(secret string) string {
	if secret == "" {
		return "<not set>"
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
