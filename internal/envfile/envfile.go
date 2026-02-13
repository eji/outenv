package envfile

import (
	"bufio"
	"os"
	"strings"
)

// Parse reads an env file and returns a map of key-value pairs.
// Supports KEY=VALUE, comments (#), empty lines, and quoted values.
func Parse(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = unquote(val)
		env[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}

// Merge reads multiple env files in order and merges them.
// Later files override earlier ones.
func Merge(paths []string) (map[string]string, error) {
	merged := make(map[string]string)
	for _, p := range paths {
		env, err := Parse(p)
		if err != nil {
			return nil, err
		}
		for k, v := range env {
			merged[k] = v
		}
	}
	return merged, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
