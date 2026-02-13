package shell

import "fmt"

// Zsh implements Shell for zsh.
type Zsh struct{}

func (Zsh) Export(key, value string) string {
	return fmt.Sprintf("export %s=%q;", key, value)
}

func (Zsh) Unset(key string) string {
	return fmt.Sprintf("unset %s;", key)
}

func (Zsh) Hook() string {
	return `_outenv_hook() {
  eval "$(outenv _apply zsh)"
}
typeset -ag precmd_functions
if [[ -z "${precmd_functions[(r)_outenv_hook]+1}" ]]; then
  precmd_functions=(_outenv_hook $precmd_functions)
fi`
}
