package main

import (
	"os"

	"pacto/internal/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	return cli.Execute()
}
