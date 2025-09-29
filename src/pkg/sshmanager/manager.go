// pkg/sshmanager/manager.go
package sshmanager

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type SSHManager struct {
	configDir string
}

func New(configDir string) *SSHManager {
	return &SSHManager{
		configDir: configDir,
	}
}

func (m *SSHManager) GenerateKeyPair(host, keyName string, bits int) error {
	keyDir := filepath.Join(m.configDir, host)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// 保存私钥
	privateKeyFile := filepath.Join(keyDir, keyName)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateKeyData := pem.EncodeToMemory(privateKeyPEM)
	if err := os.WriteFile(privateKeyFile, privateKeyData, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// 生成并保存公钥
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to generate public key: %w", err)
	}

	publicKeyFile := privateKeyFile + ".pub"
	publicKeyData := ssh.MarshalAuthorizedKey(publicKey)
	if err := os.WriteFile(publicKeyFile, publicKeyData, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

func (m *SSHManager) AddPublicKey(host, keyName, publicKeyPath string) error {
	keyDir := filepath.Join(m.configDir, host)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	targetPath := filepath.Join(keyDir, keyName+".pub")
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}
