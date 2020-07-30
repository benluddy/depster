package main

import (
	"os"

	"github.com/benluddy/depster/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
