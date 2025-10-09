package utils_test

import (
	"encoding/base64"
	"sshpky/pkg/utils"
	"testing"
)

func TestCryptoUtils_Encrypt(t *testing.T) {

	key, e3 := utils.GenerateRandomKey(24)

	t.Log("key", base64.StdEncoding.EncodeToString(key), e3)

	secrety := utils.NewCryptoUtilsWithKey(key)

	originalStr := "abc123"
	enStr, e1 := secrety.Encrypt([]byte(originalStr))
	t.Log("encrypt res:", enStr, e1)

	resB, e2 := secrety.Decrypt(enStr)
	t.Log("decode res:", string(resB), e2)
}

func TestCryptoUtils_Decode(t *testing.T) {
	secrety := utils.NewCryptoUtilsWithKey([]byte("XcpHJbc90YfempVoKNvD6cFdOb9JpIDj"))
	cc, e := secrety.Decrypt("/vSYeIr63feE6mrI4QowyIHjOkE1DMSEq3CkZaukvdr/t6P3yg==")
	t.Log("key,", string(cc), e)
}
