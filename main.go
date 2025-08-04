package main

import (
	"os"

	"github.com/acristoffers/cgen/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
