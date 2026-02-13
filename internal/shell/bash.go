package shell

import "fmt"

// Bash implements Shell for bash.
type Bash struct{}

func (Bash) Export(key, value string) string {
	return fmt.Sprintf("export %s=%q;", key, value)
}

func (Bash) Unset(key string) string {
	return fmt.Sprintf("unset %s;", key)
}

func (Bash) Hook() string {
	return `_outenv_hook() {
  local previous_exit_status=$?
  eval "$(outenv _apply bash)"
  return $previous_exit_status
}
if ! [[ "${PROMPT_COMMAND:-}" =~ _outenv_hook ]]; then
  PROMPT_COMMAND="_outenv_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi`
}
