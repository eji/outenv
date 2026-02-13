package shell

import "fmt"

// Shell defines the interface for shell-specific command generation.
type Shell interface {
	Export(key, value string) string
	Unset(key string) string
	Hook() string
}

// Get returns the Shell implementation for the given shell name.
func Get(name string) (Shell, error) {
	switch name {
	case "bash":
		return Bash{}, nil
	case "zsh":
		return Zsh{}, nil
	case "fish":
		return Fish{}, nil
	default:
		return nil, fmt.Errorf("unsupported shell: %s", name)
	}
}
