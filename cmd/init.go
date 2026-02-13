package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eji/outenv/internal/config"
)

const envTemplate = `# Add environment variables below, one per line.
# Format: KEY=VALUE
# Lines starting with # are comments.
#
# To encrypt a value: outenv encrypt "secret"
# Then use the output: SECRET_KEY=ENC:base64...
`

func RunInit() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath, err := config.EnvFilePath(wd)
	if err != nil {
		return fmt.Errorf("failed to determine env file path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if _, err := os.Stat(envPath); err == nil {
		fmt.Fprintf(os.Stderr, "env file already exists: %s\n", envPath)
		return nil
	}

	if err := os.WriteFile(envPath, []byte(envTemplate), 0o644); err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "created %s\n", envPath)
	return nil
}
