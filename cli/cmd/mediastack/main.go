package main

import (
	"os"

	"github.com/jxmullins/mediastack/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
