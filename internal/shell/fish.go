package shell

import "fmt"

// Fish implements Shell for fish.
type Fish struct{}

func (Fish) Export(key, value string) string {
	return fmt.Sprintf("set -gx %s %q;", key, value)
}

func (Fish) Unset(key string) string {
	return fmt.Sprintf("set -e %s;", key)
}

func (Fish) Hook() string {
	return `function _outenv_hook --on-variable PWD
  outenv _apply fish | source
end`
}
