package main

import (
	"fmt"
	"os"

	"github.com/eji/outenv/cmd"
)

const usage = `Usage: outenv <command> [args]

Commands:
  init          Initialize env file for current directory
  edit          Edit env file for current directory
  export        Export merged environment variables
  encrypt <value> Encrypt a value for use in env files
  hook <shell>  Print shell hook (bash, zsh, fish)
  _apply <shell> Apply environment changes (internal)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = cmd.RunInit()
	case "edit":
		err = cmd.RunEdit()
	case "export":
		err = cmd.RunExport()
	case "encrypt":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: outenv encrypt <value>")
			os.Exit(1)
		}
		err = cmd.RunEncrypt(os.Args[2])
	case "hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: outenv hook <shell>")
			os.Exit(1)
		}
		err = cmd.RunHook(os.Args[2])
	case "_apply":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: outenv _apply <shell>")
			os.Exit(1)
		}
		err = cmd.RunApply(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "outenv: %s\n", err)
		os.Exit(1)
	}
}
