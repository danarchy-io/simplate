package main

import (
	"os"

	"github.com/danarchy-io/simplate/cmd"
)

var version = "dev" // Set via ldflags during build

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
