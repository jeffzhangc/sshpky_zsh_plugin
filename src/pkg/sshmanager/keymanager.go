package sshmanager

import (
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
)

func GenerateOTP(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func SaveMFASecret(host, secret string) {

}
func GetMFASecret(host string) (string, error) {
	return GetPassword(host + "/mfa")
}

func GetPassword(host string) (string, error) {
	return "", nil
}
func SavePassword(host, password string) error {
	fmt.Println(host, "xxxxx", password)
	return nil
}
