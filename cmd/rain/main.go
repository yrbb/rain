package main

import (
	"os"

	"github.com/yrbb/rain/cmd/rain/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
