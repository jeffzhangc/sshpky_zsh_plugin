package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	DEFAULT_USENAME = "default"
	DEFAULT_KEYSIZE = 24
)

var homeDir string
var config SshpkyConfig

const configFile = "config.yaml"

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	homeDir = filepath.Join(homeDir, ".sshpky")
	initConfig()
}

func initConfig() {
	configPath := filepath.Join(homeDir, configFile)

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config = SshpkyConfig{
			Use:     DEFAULT_USENAME,
			KeySize: DEFAULT_KEYSIZE,
			Groups:  []SshpkyGroupConfig{},
		}
		SetConfig(config)
		return
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// 清理重复的group（保持第一个出现的）
	cleanDuplicateGroups()
}

func cleanDuplicateGroups() {
	seen := make(map[string]bool)
	uniqueGroups := []SshpkyGroupConfig{}
	changed := false

	for _, group := range config.Groups {
		if !seen[group.Name] {
			seen[group.Name] = true
			uniqueGroups = append(uniqueGroups, group)
		} else {
			changed = true
		}
	}

	if changed {
		config.Groups = uniqueGroups
		SetConfig(config)
	}
}

func GetConfig() SshpkyConfig {
	return config
}

func SetConfig(newConfig SshpkyConfig) {
	config = newConfig
	SaveConfig()
}

func SaveConfig() {
	out, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("Error marshaling config: %v", err)
	}

	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		err := os.MkdirAll(homeDir, 0755)
		if err != nil {
			log.Fatalf("Error creating config directory: %v", err)
		}
	}

	configPath := filepath.Join(homeDir, configFile)
	file, err := os.OpenFile(configPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(out)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		log.Fatalf("Error flushing config file: %v", err)
	}
}

func SetGroup(group SshpkyGroupConfig) {
	added := false
	for i, g := range config.Groups {
		if g.Name == group.Name {
			config.Groups[i] = group
			added = true
			break
		}
	}
	if !added {
		config.Groups = append(config.Groups, group)
	}
	SaveConfig()
}

func GetGroup(name string) (SshpkyGroupConfig, error) {
	for _, group := range config.Groups {
		if group.Name == name {
			return group, nil
		}
	}
	return SshpkyGroupConfig{}, fmt.Errorf("group '%s' not found", name)
}

func DeleteGroup(name string) error {
	if name == config.Use {
		return fmt.Errorf("cannot delete active group '%s'", name)
	}

	newGroups := []SshpkyGroupConfig{}
	found := false

	for _, group := range config.Groups {
		if group.Name == name {
			found = true
			continue
		}
		newGroups = append(newGroups, group)
	}

	if !found {
		return fmt.Errorf("group '%s' not found", name)
	}

	config.Groups = newGroups
	SaveConfig()
	return nil
}
