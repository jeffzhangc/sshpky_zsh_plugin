package config_test

import (
	"fmt"
	"sshpky/pkg/config"
	"testing"
	"time"
)

func TestSSHConfigManager_ValidateConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		configPath string
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := config.NewSSHConfigManager(tt.configPath)
			gotErr := m.ValidateConfig()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateConfig() succeeded unexpectedly")
			}
		})
	}
}

// 使用示例
func TestConfigManager(t *testing.T) {
	manager := config.NewSSHConfigManager("abc")

	// 添加配置项
	newConfig := config.SshConfigItem{
		Group:        "Production",
		Host:         "myserver",
		HostName:     "10.1.102.34",
		Port:         5522,
		User:         "root",
		IdentityFile: "~/.ssh/id_rsa",
		Desc:         "生产环境服务器",
		EditTime:     time.Now().Format("2006-01-02 15:04:05"),
		Password:     "xxx",
	}

	if err := manager.AddConfig(newConfig); err != nil {
		fmt.Printf("添加配置失败: %v\n", err)
	}

	// 查找配置项
	config, err := manager.FindConfig("myserver")
	if err != nil {
		fmt.Printf("查找配置失败: %v\n", err)
	} else {
		fmt.Printf("找到配置: %+v\n", config)
	}

	// 更新配置项
	updatedConfig := *config
	updatedConfig.Port = 5523
	if err := manager.UpdateConfig("myserver", updatedConfig); err != nil {
		fmt.Printf("更新配置失败: %v\n", err)
	}

	// 搜索配置项
	results, err := manager.SearchConfigs("production")
	if err != nil {
		fmt.Printf("搜索配置失败: %v\n", err)
	} else {
		fmt.Printf("找到 %d 个匹配的配置项\n", len(results))
	}

	// // 删除配置项
	// if err := manager.DeleteConfig("myserver"); err != nil {
	// 	fmt.Printf("删除配置失败: %v\n", err)
	// }
}
