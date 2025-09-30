package km

import (
	"fmt"
	"time"

	"github.com/keybase/go-keychain"
	"github.com/pquerna/otp/totp"
)

const (
	KEY_CHAIN_NAME = "ssh_py_default"
)

const (
	PWD_GOOGLE = "googleAuthCode"
	PWD_NORMAL = "password"
)

func GenerateOTP(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func getPasswordByType(key string, category string) (string, error) {
	pwdKey := fmt.Sprintf("%s/%s", key, category)

	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(KEY_CHAIN_NAME)
	query.SetAccount(pwdKey)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", fmt.Errorf("keychain query error: %v", err)
	}
	if len(results) != 1 {
		return "", nil
	}

	return string(results[0].Data), nil
}

func savePwdByType(key string, category string, secret string) error {
	pwdKey := fmt.Sprintf("%s/%s", key, category)

	// 首先检查是否已经存在，如果存在则更新，否则添加
	existing, err := getPasswordByType(key, category)
	if err == nil && existing != "" {
		// 更新现有项
		query := keychain.NewItem()
		query.SetSecClass(keychain.SecClassGenericPassword)
		query.SetService(KEY_CHAIN_NAME)
		query.SetAccount(pwdKey)

		updateItem := keychain.NewItem()
		updateItem.SetData([]byte(secret))

		err = keychain.UpdateItem(query, updateItem)
		if err != nil {
			return fmt.Errorf("keychain update error: %v", err)
		}
		return nil
	}

	// 添加新项
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(KEY_CHAIN_NAME)
	item.SetAccount(pwdKey)
	item.SetData([]byte(secret))
	item.SetAccessible(keychain.AccessibleWhenUnlocked)

	err = keychain.AddItem(item)
	if err != nil {
		return fmt.Errorf("keychain add error: %v", err)
	}
	return nil
}

func getHostKey(user, host string) string {
	return fmt.Sprintf("%s@%s", user, host)
}

func SaveMFASecret(user, host, secret string) error {
	return savePwdByType(getHostKey(user, host), PWD_GOOGLE, secret)
}
func GetMFASecret(user, host string) (string, error) {
	pwd, err := getPasswordByType(getHostKey(user, host), PWD_GOOGLE)
	if err != nil {
		return pwd, err
	}
	if pwd != "" && len(pwd) > 0 {
		return GenerateOTP(pwd)
	}
	return "", nil
}

func GetPassword(username, host string) (string, error) {
	return getPasswordByType(getHostKey(username, host), PWD_NORMAL)
}
func SavePassword(username, host, password string) error {
	return savePwdByType(getHostKey(username, host), PWD_NORMAL, password)
}
