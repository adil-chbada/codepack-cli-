package main

import (
	"os"

	"github.com/adil-chbada/codepack-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}