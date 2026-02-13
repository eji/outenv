package cmd

import (
	"fmt"

	"github.com/eji/outenv/internal/shell"
)

func RunHook(shellName string) error {
	sh, err := shell.Get(shellName)
	if err != nil {
		return err
	}
	fmt.Println(sh.Hook())
	return nil
}
