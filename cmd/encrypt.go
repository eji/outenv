package cmd

import (
	"fmt"

	"github.com/eji/outenv/internal/crypto"
)

func RunEncrypt(value string) error {
	key, err := crypto.LoadOrCreateKey()
	if err != nil {
		return fmt.Errorf("failed to load encryption key: %w", err)
	}

	encrypted, err := crypto.Encrypt(key, value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	fmt.Println(encrypted)
	return nil
}
