package config

import "sshpky/pkg/utils"

type KeyStoreFileManage struct {
	secret string
	crypto *utils.CryptoUtils
}

var keystoreMap map[string]IKeyM = map[string]IKeyM{}

func NewKeyStoreFileManage(secret string) IKeyM {
	if r, ok := keystoreMap[secret]; ok {
		return r
	}

	m := &KeyStoreFileManage{
		secret: secret,
		crypto: utils.NewCryptoUtilsWithKey([]byte(secret)),
	}
	keystoreMap[secret] = m
	return m
}

func (k *KeyStoreFileManage) SavePwd(connConf *SshConfigItem) {
	if connConf.Password != "" {
		encryptStr, e := k.crypto.Encrypt([]byte(connConf.Password))
		if e == nil {
			connConf.Password = encryptStr
		}
	}
	if connConf.MFASecret != "" {
		mafStr, e := k.crypto.Encrypt([]byte(connConf.MFASecret))
		if e == nil {
			connConf.MFASecret = mafStr
		}
	}
}
func (k *KeyStoreFileManage) GetPwd(connConf SshConfigItem) string {
	if connConf.Password != "" {
		dstr, e := k.crypto.Decrypt(connConf.Password)
		if e == nil {
			return string(dstr)
		}
	}
	return ""

}
func (k *KeyStoreFileManage) GetMAFSecret(connConf SshConfigItem) string {
	if connConf.MFASecret != "" {
		dstr, e := k.crypto.Decrypt(connConf.MFASecret)
		if e == nil {
			return string(dstr)
		}
	}
	return ""
}
