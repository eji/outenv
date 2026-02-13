package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/eji/outenv/internal/config"
)

func RunEdit() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath, err := config.EnvFilePath(wd)
	if err != nil {
		return fmt.Errorf("failed to determine env file path: %w", err)
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("env file not found: %s\nRun 'outenv init' first", envPath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("$EDITOR is not set")
	}

	cmd := exec.Command(editor, envPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
