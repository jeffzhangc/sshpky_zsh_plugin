package config

import "fmt"

type SecretCategory int

const (
	StoreFile SecretCategory = iota
	StoreKeyChain
)

// SshConfigItem 表示 SSH 配置文件的单个主机配置项
type SshConfigItem struct {
	Group        string // 分组名称，用于组织管理
	Host         string // 主机别名，用于 ssh 命令连接时使用
	HostName     string // 实际的主机地址或 IP
	Port         int    // SSH 端口号，默认 22
	User         string // 登录用户名
	IdentityFile string // 身份认证文件路径（私钥）
	Password     string // 登录密码（注意：明文存储密码不安全）
	MFASecret    string // mfa 密码
	EditTime     string // 最后编辑时间
	ProxyCommand string // 代理命令，用于跳板机等场景
	Desc         string // 配置项描述
	OtherParams  []string
}

// 为了在日志或输出中容易阅读，我们实现String方法
func (s SecretCategory) String() string {
	switch s {
	case StoreFile:
		return "PwStoreFiledStr"
	case StoreKeyChain:
		return "StoreKeyChain"
	default:
		return "Unknown"
	}
}

// 从字符串解析SecretCategory
func ParseSecretCategory(str string) (SecretCategory, error) {
	switch str {
	case "StoreFile":
		return StoreFile, nil
	case "StoreKeyChain":
		return StoreKeyChain, nil
	default:
		return StoreKeyChain, fmt.Errorf("invalid secret category: %s", str)
	}
}

type IKeyM interface {
	SavePwd(connConf *SshConfigItem)
	GetPwd(connConf SshConfigItem) string
	GetMAFSecret(connConf SshConfigItem) string
}

type SshpkyConfig struct {
	Use     string              `yaml:"use"`
	KeySize int                 `yaml:"keySize"`
	Groups  []SshpkyGroupConfig `yaml:"groups"`
}

type SshpkyGroupConfig struct {
	Name     string         `yaml:"name"`
	Secret   string         `yaml:"secret"`
	AutoSave bool           `yaml:"autoSave"`
	Desc     string         `yaml:"desc"`
	Category SecretCategory `yaml:"category"`
}

func (s SshpkyConfig) GetGroupNames() (res []string) {
	for _, item := range s.Groups {
		res = append(res, item.Name)
	}
	return res
}

func (sg *SshpkyGroupConfig) getKeyManager() IKeyM {
	switch sg.Category {
	case StoreFile:
		return NewKeyStoreFileManage(sg.Secret)
	case StoreKeyChain:
		return NewKeyChainManage()
	}
	return NewKeyChainManage()
}

func (sc *SshConfigItem) getGroup() *SshpkyGroupConfig {

	if sc.Group == "" {
		sc.Group = config.Use
	}

	for _, g := range config.Groups {
		if g.Name == sc.Group {
			return &g
		}
	}
	return nil
}

func (sc *SshConfigItem) GetPassword() string {
	ssm := NewSSHConfigManager("")
	dbConf, _ := ssm.FindConfig(sc.Host)
	if dbConf == nil {
		return ""
	}

	group := sc.getGroup()
	if group == nil {
		return ""
	}
	km := group.getKeyManager()
	return km.GetPwd(*sc)
}

func (sc *SshConfigItem) GetMafSecret() string {
	ssm := NewSSHConfigManager("")
	dbConf, _ := ssm.FindConfig(sc.Host)
	if dbConf == nil {
		return ""
	}

	group := sc.getGroup()
	if group == nil {
		return ""
	}
	km := group.getKeyManager()
	return km.GetMAFSecret(*sc)
}
