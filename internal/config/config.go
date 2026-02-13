package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DataDir returns the base data directory for env files.
// Uses $XDG_DATA_HOME/outenv if set, otherwise ~/.local/share/outenv.
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "outenv"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "outenv"), nil
}

// EnvFilePath returns the env file path for the given absolute directory.
// e.g. /Users/koji/project → ~/.local/share/outenv/Users/koji/project/env
func EnvFilePath(absDir string) (string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return "", err
	}
	// Remove leading "/" to make it relative for joining
	rel := strings.TrimPrefix(absDir, "/")
	return filepath.Join(dataDir, rel, "env"), nil
}

// KeyFilePath returns the path to the encryption key file.
// Uses $XDG_CONFIG_HOME/outenv/key if set, otherwise ~/.config/outenv/key.
func KeyFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "outenv", "key"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "outenv", "key"), nil
}

// AncestorEnvFiles returns env file paths from root to the given directory,
// only including paths where the env file actually exists.
func AncestorEnvFiles(absDir string) ([]string, error) {
	dataDir, err := DataDir()
	if err != nil {
		return nil, err
	}

	var dirs []string
	dir := absDir
	for {
		dirs = append(dirs, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Reverse so we go from root to current directory
	for i, j := 0, len(dirs)-1; i < j; i, j = i+1, j-1 {
		dirs[i], dirs[j] = dirs[j], dirs[i]
	}

	var files []string
	for _, d := range dirs {
		rel := strings.TrimPrefix(d, "/")
		envPath := filepath.Join(dataDir, rel, "env")
		if _, err := os.Stat(envPath); err == nil {
			files = append(files, envPath)
		}
	}
	return files, nil
}
