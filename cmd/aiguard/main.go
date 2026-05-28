package main

import (
	"os"

	"github.com/mchau/aiguard/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
