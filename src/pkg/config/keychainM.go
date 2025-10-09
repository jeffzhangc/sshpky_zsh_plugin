package config

import (
	"sshpky/pkg/km"
)

type KeyChainManage struct {
}

var keyChainM IKeyM

func init() {
	keyChainM = &KeyChainManage{}
}

func NewKeyChainManage() IKeyM {
	return keyChainM
}

func (k *KeyChainManage) SavePwd(connConf *SshConfigItem) {
	if connConf.Password != "" {
		km.SavePassword(connConf.User, connConf.Host, connConf.Password)
		connConf.Password = ""
	}
	if connConf.MFASecret != "" {
		km.SaveMFASecret(connConf.User, connConf.Host, connConf.MFASecret)
		connConf.MFASecret = ""
	}
}
func (k *KeyChainManage) GetPwd(connConf SshConfigItem) string {
	res, _ := km.GetPassword(connConf.User, connConf.Host)
	return res
}
func (k *KeyChainManage) GetMAFSecret(connConf SshConfigItem) string {
	res, _ := km.GetMFASecret(connConf.User, connConf.Host)
	return res
}
