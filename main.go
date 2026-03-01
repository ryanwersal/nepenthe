package main

import (
	"os"

	"github.com/ryanwersal/nepenthe/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
