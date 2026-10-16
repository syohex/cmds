package main

import (
	"fmt"
	"os"

	"github.com/syohex/cmds"
)

func _main() int {
	if !cmds.IsWSL() {
		fmt.Fprintf(os.Stderr, "'wsl-open' supports only WSL environment")
		return 1
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: wsl-open URL")
		return 1
	}

	url := os.Args[1]

	if err := cmds.WSLOpen(url); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open %s: %v\n", url, err)
		return 1
	}
	return 0
}

func main() {
	 os.Exit(_main())
}
