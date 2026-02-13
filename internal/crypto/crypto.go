package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eji/outenv/internal/config"
)

const (
	keySize   = 32 // AES-256
	nonceSize = 12 // GCM standard nonce size
	prefix    = "ENC:"
)

// LoadOrCreateKey loads the encryption key from the key file,
// or creates a new random key if it does not exist.
func LoadOrCreateKey() ([]byte, error) {
	keyPath, err := config.KeyFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine key file path: %w", err)
	}

	key, err := os.ReadFile(keyPath)
	if err == nil {
		if len(key) != keySize {
			return nil, fmt.Errorf("invalid key file: expected %d bytes, got %d", keySize, len(key))
		}
		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Generate new key
	key = make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(keyPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}

	return key, nil
}

// LoadKey loads the encryption key from the key file.
// Returns an error if the key file does not exist.
func LoadKey() ([]byte, error) {
	keyPath, err := config.KeyFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine key file path: %w", err)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	if len(key) != keySize {
		return nil, fmt.Errorf("invalid key file: expected %d bytes, got %d", keySize, len(key))
	}

	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a string
// in the format "ENC:base64(nonce||ciphertext||tag)".
func Encrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal appends ciphertext+tag to nonce
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return prefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts an "ENC:base64..." string using AES-256-GCM
// and returns the plaintext.
func Decrypt(key []byte, encoded string) (string, error) {
	if !IsEncrypted(encoded) {
		return "", fmt.Errorf("value is not encrypted (missing %s prefix)", prefix)
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(encoded, prefix))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted returns true if the value has the "ENC:" prefix.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, prefix)
}
